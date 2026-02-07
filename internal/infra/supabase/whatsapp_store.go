package supabase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	supa "github.com/nedpals/supabase-go"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	walog "go.mau.fi/whatsmeow/util/log"
)

type SupabaseContainer struct {
	client *supa.Client
	log    walog.Logger
}

func NewSupabaseContainer(supabaseURL, supabaseKey string, logger walog.Logger) (*SupabaseContainer, error) {
	client := supa.CreateClient(supabaseURL, supabaseKey)

	return &SupabaseContainer{
		client: client,
		log:    logger,
	}, nil
}

func (c *SupabaseContainer) GetAllDevices(ctx context.Context) ([]*store.Device, error) {
	var devices []*store.Device

	// Get credentials from Supabase
	var results []map[string]interface{}
	err := c.client.DB.From("whatsapp_sessions").
		Select("*").
		Execute(&results)

	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if len(results) == 0 {
		return devices, nil
	}

	for _, row := range results {
		sessionDataStr := row["session_data"].(string)
		device, err := unmarshalDeviceFromJSON([]byte(sessionDataStr))
		if err != nil {
			c.log.Warnf("Failed to unmarshal device: %v", err)
			continue
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (c *SupabaseContainer) GetFirstDevice(ctx context.Context) (*store.Device, error) {
	devices, err := c.GetAllDevices(ctx)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, sql.ErrNoRows
	}
	return devices[0], nil
}

func (c *SupabaseContainer) NewDevice() *store.Device {
	jid := types.NewJID(uuid.New().String(), types.DefaultUserServer)
	device := &store.Device{
		ID: &jid,
	}
	return device
}

func (c *SupabaseContainer) PutDevice(ctx context.Context, device *store.Device) error {
	deviceID := device.ID.String()
	c.log.Debugf("Attempting to save device %s to Supabase", deviceID)

	// Create a clean device map for JSON serialization
	deviceMap := map[string]interface{}{
		"ID":             deviceID,
		"RegistrationID": device.RegistrationID,
		"Platform":       device.Platform,
		"BusinessName":   device.BusinessName,
		"PushName":       device.PushName,
	}

	// Convert device to JSON
	deviceJSON, err := json.Marshal(deviceMap)
	if err != nil {
		return fmt.Errorf("failed to marshal device: %w", err)
	}

	// Try using Upsert with proper conflict resolution
	c.log.Debugf("Attempting UPSERT for device %s", deviceID)

	var results []map[string]interface{}
	err = c.client.DB.From("whatsapp_sessions").
		Upsert(map[string]interface{}{
			"device_id":    deviceID,
			"session_data": string(deviceJSON),
			"updated_at":   time.Now(),
		}).
		Execute(&results)

	if err != nil {
		c.log.Errorf("UPSERT failed for device %s: %v", deviceID, err)

		// Fallback: Try DELETE + INSERT with retry logic
		c.log.Warnf("UPSERT failed, trying DELETE + INSERT fallback for device %s", deviceID)

		// Delete with retry
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			err = c.client.DB.From("whatsapp_sessions").
				Delete().
				Eq("device_id", deviceID).
				Execute(&results)
			if err == nil {
				break
			}
			c.log.Debugf("Delete attempt %d failed for device %s: %v", i+1, deviceID, err)
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}

		if err != nil {
			return fmt.Errorf("failed to delete existing device after %d retries: %w", maxRetries, err)
		}

		// Insert with retry
		for i := 0; i < maxRetries; i++ {
			err = c.client.DB.From("whatsapp_sessions").
				Insert(map[string]interface{}{
					"device_id":    deviceID,
					"session_data": string(deviceJSON),
					"updated_at":   time.Now(),
				}).
				Execute(&results)
			if err == nil {
				break
			}
			c.log.Debugf("Insert attempt %d failed for device %s: %v", i+1, deviceID, err)
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}

		if err != nil {
			return fmt.Errorf("failed to insert device after %d retries: %w", maxRetries, err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to save device: %w", err)
	}

	c.log.Infof("Device %s saved successfully to Supabase", deviceID)
	return nil
}

func (c *SupabaseContainer) DeleteDevice(ctx context.Context, id types.JID) error {
	var results []map[string]interface{}
	err := c.client.DB.From("whatsapp_sessions").
		Delete().
		Eq("device_id", id.String()).
		Execute(&results)

	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	return nil
}

// unmarshalDeviceFromJSON handles device unmarshaling with proper field reconstruction
func unmarshalDeviceFromJSON(data []byte) (*store.Device, error) {
	// Create a temporary map to extract all fields
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	// Create a new device
	device := &store.Device{}

	// Extract basic string fields
	if idStr, ok := temp["ID"].(string); ok && idStr != "" {
		jid, err := types.ParseJID(idStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse device ID: %w", err)
		}
		device.ID = &jid
	}

	if regID, ok := temp["RegistrationID"].(float64); ok {
		device.RegistrationID = uint32(regID)
	}

	if platform, ok := temp["Platform"].(string); ok {
		device.Platform = platform
	}

	if businessName, ok := temp["BusinessName"].(string); ok {
		device.BusinessName = businessName
	}

	if pushName, ok := temp["PushName"].(string); ok {
		device.PushName = pushName
	}

	return device, nil
}

// Helper method to restore device from Supabase by device ID
func (c *SupabaseContainer) RestoreDeviceFromSupabase(ctx context.Context, deviceID string) (*store.Device, error) {
	var results []map[string]interface{}
	err := c.client.DB.From("whatsapp_sessions").
		Select("*").
		Eq("device_id", deviceID).
		Execute(&results)

	if err != nil {
		return nil, fmt.Errorf("failed to get session data: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no session data found for device %s", deviceID)
	}

	sessionDataStr := results[0]["session_data"].(string)
	device, err := unmarshalDeviceFromJSON([]byte(sessionDataStr))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return device, nil
}

// SimpleBackupToSupabase provides a way to backup any device to Supabase
func SimpleBackupToSupabase(supabaseURL, supabaseKey string, device *store.Device, logger walog.Logger) error {
	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("supabase credentials not provided")
	}

	container, err := NewSupabaseContainer(supabaseURL, supabaseKey, logger)
	if err != nil {
		return fmt.Errorf("failed to create supabase container: %w", err)
	}

	ctx := context.Background()
	err = container.PutDevice(ctx, device)
	if err != nil {
		return fmt.Errorf("failed to backup device to supabase: %w", err)
	}

	logger.Infof("Device successfully backed up to Supabase")
	return nil
}

// SimpleRestoreFromSupabase provides a way to restore any device from Supabase
func SimpleRestoreFromSupabase(supabaseURL, supabaseKey string, deviceID string, logger walog.Logger) (*store.Device, error) {
	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("supabase credentials not provided")
	}

	container, err := NewSupabaseContainer(supabaseURL, supabaseKey, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase container: %w", err)
	}

	ctx := context.Background()
	device, err := container.RestoreDeviceFromSupabase(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to restore device from supabase: %w", err)
	}

	logger.Infof("Device successfully restored from Supabase")
	return device, nil
}

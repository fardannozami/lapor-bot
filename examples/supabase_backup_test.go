package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fardannozami/whatsapp-gateway/internal/infra/supabase"
	"go.mau.fi/whatsmeow/store"
	walog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// Example: Backup and restore WhatsApp credentials using Supabase

	// Initialize logger
	logger := walog.Stdout("Test", "INFO", true)

	// Supabase configuration (replace with your actual values)
	supabaseURL := "https://your-project.supabase.co"
	supabaseKey := "your-anon-key"

	ctx := context.Background()

	// Example 1: Backup a device to Supabase
	fmt.Println("=== Backup Example ===")

	// Create a sample device (in real usage, this comes from logged-in client)
	device := &store.Device{
		// This would be populated from actual device after login
		// For demo purposes, we'll create a simple device
	}

	err := supabase.SimpleBackupToSupabase(supabaseURL, supabaseKey, device, logger)
	if err != nil {
		log.Printf("Backup failed: %v", err)
	} else {
		fmt.Println("✅ Device backed up successfully")
	}

	// Example 2: Restore a device from Supabase
	fmt.Println("\n=== Restore Example ===")

	deviceID := "sample-device-id" // Replace with actual device ID
	restoredDevice, err := supabase.SimpleRestoreFromSupabase(supabaseURL, supabaseKey, deviceID, logger)
	if err != nil {
		log.Printf("Restore failed: %v", err)
	} else {
		fmt.Printf("✅ Device restored successfully: %+v\n", restoredDevice.ID)
	}

	// Example 3: Using SupabaseContainer directly
	fmt.Println("\n=== Direct Container Usage ===")

	container, err := supabase.NewSupabaseContainer(supabaseURL, supabaseKey, logger)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	// Get all devices
	devices, err := container.GetAllDevices(ctx)
	if err != nil {
		log.Printf("Failed to get devices: %v", err)
	} else {
		fmt.Printf("Found %d devices in Supabase\n", len(devices))
		for i, device := range devices {
			fmt.Printf("  Device %d: %s\n", i+1, device.ID.String())
		}
	}

	// Create a new device
	newDevice := container.NewDevice()
	fmt.Printf("Created new device with ID: %s\n", newDevice.ID.String())

	// Save the device
	err = container.PutDevice(ctx, newDevice)
	if err != nil {
		log.Printf("Failed to save device: %v", err)
	} else {
		fmt.Println("✅ New device saved successfully")
	}

	fmt.Println("\n=== Integration Tips ===")
	fmt.Println("1. Set SUPABASE_URL and SUPABASE_KEY in your .env file")
	fmt.Println("2. Run the SQL migration to create the necessary tables")
	fmt.Println("3. The bot will automatically backup/restore credentials")
	fmt.Println("4. You can manually backup/restore using the helper functions")
}

package wa

import (
	"context"
	"fmt"
	"os"

	"github.com/fardannozami/whatsapp-gateway/internal/infra/supabase"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	walog "go.mau.fi/whatsmeow/util/log"
	_ "modernc.org/sqlite"
)

type Service struct {
	client         *whatsmeow.Client
	dbBasePath     string
	log            walog.Logger
	messageHandler func(ctx context.Context, client *whatsmeow.Client, evt *events.Message)
	supabaseURL    string
	supabaseKey    string
}

func NewService(dbBasePath string, logger walog.Logger, supabaseURL, supabaseKey string) *Service {
	return &Service{
		dbBasePath:  dbBasePath,
		log:         logger,
		supabaseURL: supabaseURL,
		supabaseKey: supabaseKey,
	}
}

func (s *Service) Initialize(ctx context.Context) error {
	var device *store.Device
	var sqlContainer *sqlstore.Container

	// Always initialize SQLite container first
	dbAddress := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", s.dbBasePath)
	var err error
	sqlContainer, err = sqlstore.New(context.Background(), "sqlite", dbAddress, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Try to get device from SQLite first
	sqliteDevices, err := sqlContainer.GetAllDevices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get devices from SQLite: %w", err)
	}

	// If device exists in SQLite, use it
	if len(sqliteDevices) > 0 {
		device = sqliteDevices[0]
		s.log.Infof("Device loaded from SQLite")

		// Note: Auto-backup will be triggered after successful connection
		// in the event handler, not here during initialization
	} else {
		// No device in SQLite, try to restore from Supabase
		if s.supabaseURL != "" && s.supabaseKey != "" {
			supabaseContainer, err := supabase.NewSupabaseContainer(s.supabaseURL, s.supabaseKey, s.log)
			if err == nil {
				supabaseDevices, err := supabaseContainer.GetAllDevices(ctx)
				if err == nil && len(supabaseDevices) > 0 {
					// We found a device in Supabase, but we can't restore it directly
					// because cryptographic keys are missing. Instead, we'll create a new device
					// and let the user login again.
					s.log.Warnf("Found device in Supabase but cannot restore due to missing cryptographic keys")
					s.log.Infof("Please login again to recreate the device")
				}
			}
		}

		// Create new device
		device = sqlContainer.NewDevice()
		s.log.Infof("New device created")
	}

	// Initialize client
	s.client = whatsmeow.NewClient(device, s.log)
	s.registerEventHandlers()

	return nil
}

func (s *Service) Connect() error {
	if s.client == nil {
		return fmt.Errorf("client not initialized")
	}
	if s.client.IsConnected() {
		return nil
	}
	return s.client.Connect()
}

func (s *Service) Disconnect() {
	if s.client != nil {
		s.client.Disconnect()
	}
}

func (s *Service) SetMessageHandler(handler func(ctx context.Context, client *whatsmeow.Client, evt *events.Message)) {
	s.messageHandler = handler
}

func (s *Service) registerEventHandlers() {
	s.client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if s.messageHandler != nil {
				go s.messageHandler(context.Background(), s.client, v)
			}
		case *events.Connected:
			s.log.Infof("WhatsApp connected successfully")
			// Temporarily disable auto-save to test manual backup
			s.log.Infof("Auto-save to Supabase temporarily disabled for testing")
		// TODO: Re-enable after fixing duplicate key issue
		/*
			if s.supabaseURL != "" && s.supabaseKey != "" && s.IsLoggedIn() {
				s.log.Infof("Starting auto-save to Supabase...")
				go func() {
					// Add delay to avoid race conditions
					time.Sleep(2 * time.Second)
					err := s.SaveDeviceToSupabase(context.Background())
					if err != nil {
						s.log.Warnf("Failed to auto-save device to Supabase: %v", err)
					} else {
						s.log.Infof("Device saved to Supabase")
					}
				}()
			}
		*/
		case *events.LoggedOut:
			s.log.Infof("WhatsApp logged out")
		}
	})
}

func (s *Service) GetClient() *whatsmeow.Client {
	return s.client
}

func (s *Service) IsLoggedIn() bool {
	return s.client.Store.ID != nil
}

func (s *Service) Pair(phone string) (string, error) {
	if s.IsLoggedIn() {
		return "", fmt.Errorf("already logged in")
	}

	// Ensure connected before pairing
	if !s.client.IsConnected() {
		return "", fmt.Errorf("client not connected")
	}

	// PairPhone(phone, showPushNotification, clientType, clientDisplayName)
	code, err := s.client.PairPhone(context.Background(), phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) PrintQR() {
	if s.client.Store.ID == nil {
		qrChan, _ := s.client.GetQRChannel(context.Background())
		err := s.client.Connect()
		if err != nil {
			fmt.Println("Failed to connect for QR:", err)
			return
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR Code:", evt.Code)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	}
}

func (s *Service) SaveDeviceToSupabase(ctx context.Context) error {
	if s.supabaseURL == "" || s.supabaseKey == "" {
		return fmt.Errorf("supabase credentials not provided")
	}

	if s.client == nil {
		return fmt.Errorf("client is nil")
	}

	if s.client.Store == nil {
		return fmt.Errorf("client store is nil")
	}

	supabaseContainer, err := supabase.NewSupabaseContainer(s.supabaseURL, s.supabaseKey, s.log)
	if err != nil {
		return fmt.Errorf("failed to create supabase container: %w", err)
	}

	err = supabaseContainer.PutDevice(ctx, s.client.Store)
	if err != nil {
		return fmt.Errorf("failed to save device to supabase: %w", err)
	}

	s.log.Infof("Device saved to Supabase")
	return nil
}

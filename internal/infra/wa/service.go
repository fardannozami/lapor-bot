package wa

import (
	"context"
	"fmt"

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
}

func NewService(dbBasePath string, logger walog.Logger) *Service {
	return &Service{
		dbBasePath: dbBasePath,
		log:        logger,
	}
}

func (s *Service) Connect(ctx context.Context) error {
	// Initialize the database container
	// Use slightly different pragmas or same as main? best to share.
	// whatsmeow uses its own connection. WAL mode persists on the DB file, so once enabled by one, it sticks.
	// But adding busy_timeout is good practice.
	// Note: sqlstore.New takes a dialect and address. formatting the address with pragmas.
	dbAddress := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", s.dbBasePath)
	container, err := sqlstore.New(context.Background(), "sqlite", dbAddress, s.log)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Get the first device (default session)
	// Get all devices
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	// If no device exists, create a new one
	var device *store.Device
	if len(devices) > 0 {
		device = devices[0]
	} else {
		device = container.NewDevice()
	}

	// Initialize the client
	s.client = whatsmeow.NewClient(device, s.log)
	s.registerEventHandlers()

	// Connect to WhatsApp
	if s.client.Store.ID == nil {
		// If not logged in, wait for QR scan (handled by main)
		// We just connect here
	}

	err = s.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
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
				// In a real terminal, use a QR library to print to console
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	}
}

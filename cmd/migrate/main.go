package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fardannozami/whatsapp-gateway/internal/infra/supabase"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	walog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	fmt.Println("🔄 WhatsApp DB to Supabase Migration Tool")
	fmt.Println("==========================================")

	// Get configuration
	sqlitePath := getEnv("SQLITE_PATH", "./data/whatsapp.db")
	supabaseURL := getEnv("SUPABASE_URL", "")
	supabaseKey := getEnv("SUPABASE_KEY", "")

	// Validate configuration
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("❌ SUPABASE_URL and SUPABASE_KEY must be set in environment variables")
	}

	// Check if SQLite file exists
	if _, err := os.Stat(sqlitePath); os.IsNotExist(err) {
		log.Fatalf("❌ SQLite file not found: %s", sqlitePath)
	}

	fmt.Printf("📁 SQLite Path: %s\n", sqlitePath)
	fmt.Printf("☁️  Supabase URL: %s\n", supabaseURL)

	// Initialize logger
	logger := walog.Stdout("Migration", "INFO", true)

	// Create Supabase container
	container, err := supabase.NewSupabaseContainer(supabaseURL, supabaseKey, logger)
	if err != nil {
		log.Fatalf("❌ Failed to create Supabase container: %v", err)
	}

	ctx := context.Background()

	// Step 1: Read from SQLite using whatsmeow sqlstore
	fmt.Println("\n📖 Step 1: Reading devices from SQLite...")
	sqliteDevices, err := readDevicesUsingSQLStore(sqlitePath, logger)
	if err != nil {
		log.Fatalf("❌ Failed to read from SQLite: %v", err)
	}

	if len(sqliteDevices) == 0 {
		fmt.Println("ℹ️  No devices found in SQLite database")
		return
	}

	fmt.Printf("✅ Found %d device(s) in SQLite\n", len(sqliteDevices))

	// Step 2: Backup to Supabase
	fmt.Println("\n💾 Step 2: Backing up devices to Supabase...")
	successCount := 0

	for i, deviceInterface := range sqliteDevices {
		device := deviceInterface.(*store.Device)
		fmt.Printf("   Migrating device %d/%d: %s\n", i+1, len(sqliteDevices), device.ID.String())

		err := container.PutDevice(ctx, device)
		if err != nil {
			fmt.Printf("   ❌ Failed to backup device %s: %v\n", device.ID.String(), err)
			continue
		}
		successCount++
		fmt.Printf("   ✅ Successfully backed up device %s\n", device.ID.String())
	}

	// Step 3: Verification
	fmt.Println("\n🔍 Step 3: Verifying migration...")
	supabaseDevices, err := container.GetAllDevices(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to verify migration: %v", err)
	}

	fmt.Printf("✅ Verification complete. %d devices found in Supabase\n", len(supabaseDevices))

	// Summary
	fmt.Println("\n📊 Migration Summary")
	fmt.Println("=====================")
	fmt.Printf("Total devices in SQLite: %d\n", len(sqliteDevices))
	fmt.Printf("Successfully migrated: %d\n", successCount)
	fmt.Printf("Failed migrations: %d\n", len(sqliteDevices)-successCount)
	fmt.Printf("Total devices in Supabase: %d\n", len(supabaseDevices))

	if successCount == len(sqliteDevices) {
		fmt.Println("\n🎉 Migration completed successfully!")
		fmt.Println("💡 You can now safely backup the SQLite file and use Supabase as your credential storage.")
		fmt.Println("💡 To test: 1) Stop the bot, 2) Delete/rename whatsapp.db files, 3) Restart bot")
	} else {
		fmt.Println("\n⚠️  Migration completed with some errors.")
		fmt.Println("💡 Check the logs above for details on failed migrations.")
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// readDevicesUsingSQLStore uses the actual whatsmeow sqlstore to read devices
func readDevicesUsingSQLStore(dbPath string, logger walog.Logger) ([]interface{}, error) {
	ctx := context.Background()

	// Create the database connection string with proper pragmas
	dbAddress := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", dbPath)

	// Create sqlstore container
	sqlContainer, err := sqlstore.New(ctx, "sqlite", dbAddress, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlstore: %w", err)
	}

	// Get all devices from SQLite
	devices, err := sqlContainer.GetAllDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices from sqlstore: %w", err)
	}

	// Convert to interface slice for compatibility
	deviceInterfaces := make([]interface{}, len(devices))
	for i, device := range devices {
		deviceInterfaces[i] = device
	}

	return deviceInterfaces, nil
}

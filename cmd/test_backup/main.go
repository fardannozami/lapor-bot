package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fardannozami/whatsapp-gateway/internal/infra/supabase"
	"go.mau.fi/whatsmeow/store"
	walog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	fmt.Println("🔧 Manual Backup Test Tool")
	fmt.Println("==========================")

	// Get Supabase credentials from environment
	supabaseURL := getEnv("SUPABASE_URL", "")
	supabaseKey := getEnv("SUPABASE_KEY", "")

	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("❌ SUPABASE_URL and SUPABASE_KEY must be set")
	}

	fmt.Printf("☁️  Supabase URL: %s\n", supabaseURL)

	// Initialize logger
	logger := walog.Stdout("Test", "INFO", true)

	// Create Supabase container
	container, err := supabase.NewSupabaseContainer(supabaseURL, supabaseKey, logger)
	if err != nil {
		log.Fatalf("❌ Failed to create Supabase container: %v", err)
	}

	// Create a test device (in real usage, this comes from logged-in client)
	device := &store.Device{
		// This would be populated from actual device after login
		// For demo purposes, we'll create a simple device
	}

	fmt.Println("📦 Testing backup with test device...")

	// Test backup
	err = container.PutDevice(context.Background(), device)
	if err != nil {
		fmt.Printf("❌ Backup failed: %v\n", err)
		return
	}

	fmt.Println("✅ Backup successful!")

	// Test retrieval
	fmt.Println("\n🔍 Testing retrieval...")
	devices, err := container.GetAllDevices(context.Background())
	if err != nil {
		fmt.Printf("❌ Retrieval failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Retrieved %d devices\n", len(devices))

	// Test duplicate backup
	fmt.Println("\n🔄 Testing duplicate backup...")
	err = container.PutDevice(context.Background(), device)
	if err != nil {
		fmt.Printf("❌ Duplicate backup failed: %v\n", err)
	} else {
		fmt.Println("✅ Duplicate backup successful!")
	}

	fmt.Println("\n🎉 Manual backup test completed!")
	fmt.Println("💡 If this works, the issue might be in the auto-save timing")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

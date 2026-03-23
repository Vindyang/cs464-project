package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/vindyang/cs464-project/backend/services/shared/database"
)

func main() {
	// Load .env file
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Connect to database
	db, err := database.ConnectFromEnv()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Database connection successful!")

	// Test query - count tables
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Printf("✅ Found %d tables in public schema\n", count)

	// List tables
	var tables []string
	err = db.Select(&tables, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}

	fmt.Println("\n📋 Tables:")
	for _, table := range tables {
		fmt.Printf("  - %s\n", table)
	}

	fmt.Println("\n🎉 Database verification complete!")
}

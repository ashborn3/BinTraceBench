package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/ashborn3/BinTraceBench/internal/auth"
	"github.com/ashborn3/BinTraceBench/internal/config"
	"github.com/ashborn3/BinTraceBench/internal/database"
	"golang.org/x/term"
)

func main() {
	fmt.Println("BinTraceBench Admin Tool")
	fmt.Println("This tool helps you create admin users and manage the database.")
	fmt.Println()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	dbFactory := database.NewFactory()
	db, err := dbFactory.Create(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Choose an option:")
		fmt.Println("1. Create new user")
		fmt.Println("2. Create admin user")
		fmt.Println("3. List users")
		fmt.Println("4. Exit")
		fmt.Print("Enter choice (1-4): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			createUser(db, "user")
		case "2":
			createUser(db, "admin")
		case "3":
			listUsers(db)
		case "4":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
		fmt.Println()
	}
}

func createUser(db database.Database, role string) {
	reader := bufio.NewReader(os.Stdin)
	authService := &auth.Service{}

	fmt.Printf("Creating new %s user:\n", role)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}
	existing, err := db.GetUserByUsername(username)
	if err != nil {
		fmt.Printf("Error checking existing user: %v\n", err)
		return
	}
	if existing != nil {
		fmt.Println("User already exists.")
		return
	}
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if email == "" {
		fmt.Println("Email cannot be empty.")
		return
	}
	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		return
	}
	fmt.Println() // New line after password input

	if len(password) < 6 {
		fmt.Println("Password must be at least 6 characters long.")
		return
	}

	hashedPassword, err := authService.HashPassword(string(password))
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		return
	}

	user := &database.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Role:     role,
	}

	if err := db.CreateUser(user); err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return
	}

	fmt.Printf("✅ User '%s' created successfully with ID %d\n", username, user.ID)
}

func listUsers(db database.Database) {
	// implement a proper GetAllUsers method
	fmt.Println("Listing users:")
	fmt.Println("Note: This admin tool doesn't implement user listing yet.")
	fmt.Println("You can check the database directly or implement the GetAllUsers method.")
}

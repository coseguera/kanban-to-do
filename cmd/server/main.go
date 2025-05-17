// Package main is the entry point for the application
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/coseguera/kanban-to-do/internal/auth"
	"github.com/coseguera/kanban-to-do/internal/handlers"
	"github.com/coseguera/kanban-to-do/internal/templates"
	"github.com/coseguera/kanban-to-do/pkg/microsoft"
)

func main() {
	// Get Microsoft OAuth configuration from environment variables
	clientID := os.Getenv("MS_CLIENT_ID")
	clientSecret := os.Getenv("MS_CLIENT_SECRET")

	// Check if environment variables are set
	if clientID == "" || clientSecret == "" {
		log.Fatal("Please set MS_CLIENT_ID and MS_CLIENT_SECRET environment variables")
	}

	// Create directories if they don't exist
	if err := os.MkdirAll("templates", 0755); err != nil {
		log.Fatalf("Failed to create templates directory: %v", err)
	}

	if err := os.MkdirAll("certs", 0755); err != nil {
		log.Fatalf("Failed to create certs directory: %v", err)
	}

	// Load templates
	if err := templates.LoadTemplates("templates"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Create Microsoft client
	msConfig := microsoft.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  "https://localhost:8443/auth/callback",
		AuthURL:      "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize",
		TokenURL:     "https://login.microsoftonline.com/consumers/oauth2/v2.0/token",
		Scope:        "offline_access User.Read Tasks.ReadWrite",
		GraphURL:     "https://graph.microsoft.com/v1.0/me/todo/lists",
	}
	msClient := microsoft.NewClient(msConfig)

	// Create session manager
	sessionManager := auth.NewSessionManager(msClient)

	// Create handlers
	h := handlers.NewHandler(msClient, sessionManager)

	// Set up routes
	http.HandleFunc("/", h.HomeHandler)
	http.HandleFunc("/login", h.LoginHandler)
	http.HandleFunc("/auth/callback", h.CallbackHandler)
	http.HandleFunc("/todoLists", h.TodoListsHandler)
	http.HandleFunc("/list/", h.TasksHandler)                               // New route for tasks
	http.HandleFunc("/api/updateTask", h.UpdateTaskHandler)                 // API endpoint for updating tasks
	http.HandleFunc("/api/toggleImportance", h.ToggleTaskImportanceHandler) // API endpoint for toggling importance
	http.HandleFunc("/api/getTaskDetails", h.GetTaskDetailsHandler)         // API endpoint for getting task details
	http.HandleFunc("/api/updateTaskDetails", h.UpdateTaskDetailsHandler)   // API endpoint for updating task details
	http.HandleFunc("/api/createTask", h.CreateTaskHandler)                 // API endpoint for creating a new task
	http.HandleFunc("/logout", h.LogoutHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Set up HTTPS server
	certFile := filepath.Join("certs", "server.crt")
	keyFile := filepath.Join("certs", "server.key")

	// Check if cert/key files exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Println("Certificate files not found. Please generate a self-signed certificate using:")
		log.Println("openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj '/CN=localhost'")
		log.Fatal("Certificate files required for HTTPS")
	}

	log.Println("Starting HTTPS server on :8443...")
	log.Fatal(http.ListenAndServeTLS(":8443", certFile, keyFile, nil))
}

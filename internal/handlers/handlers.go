// Package handlers provides HTTP handlers for the application
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/coseguera/kanban-to-do/internal/auth"
	"github.com/coseguera/kanban-to-do/internal/templates"
	"github.com/coseguera/kanban-to-do/pkg/microsoft"
)

// Handler contains the dependencies for the HTTP handlers
type Handler struct {
	Client         *microsoft.Client
	SessionManager *auth.SessionManager
}

// NewHandler creates a new Handler
func NewHandler(client *microsoft.Client, sessionManager *auth.SessionManager) *Handler {
	return &Handler{
		Client:         client,
		SessionManager: sessionManager,
	}
}

// HomeHandler handles the home page
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates.Templates["home"]
	if tmpl == nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// LoginHandler handles the login request
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a state parameter to prevent CSRF
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create the authorization URL
	authRequestURL := h.Client.GetAuthURL(state)

	http.Redirect(w, r, authRequestURL, http.StatusFound)
}

// CallbackHandler handles the OAuth callback
func (h *Handler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in callback", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenResp, err := h.Client.ExchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Error exchanging code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a session
	sessionID, err := h.SessionManager.CreateSession(tokenResp)
	if err != nil {
		http.Error(w, "Error creating session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set a cookie with the session ID
	auth.SetSessionCookie(w, sessionID)

	// Redirect to the to-do lists page
	http.Redirect(w, r, "/todoLists", http.StatusFound)
}

// TodoListsHandler handles the to-do lists page
func (h *Handler) TodoListsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Get the to-do lists
	todoLists, err := h.Client.GetTodoLists(session.AccessToken)
	if err != nil {
		http.Error(w, "Error getting to-do lists: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl := templates.Templates["todoLists"]
	if tmpl == nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, todoLists)
}

// LogoutHandler handles the logout request
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err == nil {
		// Delete the session
		h.SessionManager.DeleteSession(sessionID)
	}

	// Clear the session cookie
	auth.ClearSessionCookie(w)

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// Package handlers provides HTTP handlers for the application
package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coseguera/kanban-to-do/internal/auth"
	"github.com/coseguera/kanban-to-do/internal/models"
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

// TasksHandler handles the tasks page for a specific list
func (h *Handler) TasksHandler(w http.ResponseWriter, r *http.Request) {
	// Extract list ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid list ID", http.StatusBadRequest)
		return
	}
	listID := parts[2]

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

	// Get the list details
	list, err := h.Client.GetListDetails(session.AccessToken, listID)
	if err != nil {
		http.Error(w, "Error getting list details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the tasks
	taskResp, err := h.Client.GetListTasks(session.AccessToken, listID)
	if err != nil {
		http.Error(w, "Error getting tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize columns for Kanban view
	notStartedTasks := []models.TaskDisplay{}
	doingTasks := []models.TaskDisplay{}
	doneTasks := []models.TaskDisplay{}

	// Convert tasks to display format and organize into columns
	for _, task := range taskResp.Value {
		taskDisplay := models.TaskDisplay{
			ID:         task.ID,
			Title:      task.Title,
			Status:     task.Status == "completed",
			Importance: task.Importance == "high",
			Categories: task.Categories,
		}

		// Format the due date if present
		if task.DueDateTime != nil {
			// Parse the due date
			t, err := time.Parse(time.RFC3339, task.DueDateTime.DateTime)
			if err == nil {
				// Format as a more readable date
				taskDisplay.DueDateTime = t.Format("Jan 2, 2006")
			} else {
				taskDisplay.DueDateTime = task.DueDateTime.DateTime
			}
		}

		// Determine which column this task belongs in
		if task.Status == "completed" {
			doneTasks = append(doneTasks, taskDisplay)
		} else {
			// Check if task has "Doing" category
			hasDoing := false
			for _, category := range task.Categories {
				if strings.EqualFold(category, "Doing") {
					hasDoing = true
					break
				}
			}

			if hasDoing {
				doingTasks = append(doingTasks, taskDisplay)
			} else {
				notStartedTasks = append(notStartedTasks, taskDisplay)
			}
		}
	}

	// Create TaskViewModel with Kanban columns
	taskViewModel := models.TaskViewModel{
		ListID:   listID,
		ListName: list.DisplayName,
		Columns: []models.KanbanColumn{
			{Title: "Not Started", Tasks: notStartedTasks},
			{Title: "Doing", Tasks: doingTasks},
			{Title: "Done", Tasks: doneTasks},
		},
	}

	// Render the template
	tmpl := templates.Templates["tasks"]
	if tmpl == nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, taskViewModel)
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

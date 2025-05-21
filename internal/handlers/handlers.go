// Copyright (c) 2025 Carlos Oseguera (@coseguera)
// This code is licensed under a dual-license model.
// See LICENSE.md for more information.

// Package handlers provides HTTP handlers for the application
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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

// UpdateTaskHandler handles updating task status and categories when dragged between columns
func (h *Handler) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Debug the incoming form values
	log.Printf("Form values received: %+v", r.Form)

	listID := r.FormValue("listId")
	taskID := r.FormValue("taskId")
	column := r.FormValue("column")

	log.Printf("Parsed values - listID: '%s', taskID: '%s', column: '%s'", listID, taskID, column)

	if listID == "" || taskID == "" || column == "" {
		http.Error(w, fmt.Sprintf("Missing required parameters (listId: %s, taskId: %s, column: %s)",
			listID, taskID, column), http.StatusBadRequest)
		return
	}

	// Get the task to preserve any existing categories
	taskResp, err := h.Client.GetListTasks(session.AccessToken, listID)
	if err != nil {
		http.Error(w, "Error fetching task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the task in the response
	var targetTask *models.Task
	for i := range taskResp.Value {
		if taskResp.Value[i].ID == taskID {
			targetTask = &taskResp.Value[i]
			break
		}
	}

	if targetTask == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get current categories and filter out any "Doing" category
	categories := []string{}
	for _, cat := range targetTask.Categories {
		if !strings.EqualFold(cat, "Doing") {
			categories = append(categories, cat)
		}
	}

	// Determine new status and categories based on the target column
	status := "notStarted"
	if column == "Done" {
		status = "completed"
	} else if column == "Doing" {
		status = "notStarted"
		categories = append(categories, "Doing")
	} else if column == "Not Started" {
		status = "notStarted"
		// No need to add categories, as we already filtered out "Doing"
	} else {
		// Unknown column
		log.Printf("Warning: Unknown column name received: %s", column)
	}

	// Update the task
	if err := h.Client.UpdateTaskStatus(session.AccessToken, listID, taskID, status, categories); err != nil {
		http.Error(w, "Error updating task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task updated successfully"))
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

// ToggleTaskImportanceHandler handles toggling a task's importance
func (h *Handler) ToggleTaskImportanceHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the request body
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Debug the incoming form values
	log.Printf("Form values received: %+v", r.Form)

	listID := r.FormValue("listId")
	taskID := r.FormValue("taskId")
	isImportant := r.FormValue("isImportant")

	log.Printf("Parsed values - listID: '%s', taskID: '%s', isImportant: '%s'", listID, taskID, isImportant)

	if listID == "" || taskID == "" || isImportant == "" {
		http.Error(w, fmt.Sprintf("Missing required parameters (listId: %s, taskId: %s, isImportant: %s)",
			listID, taskID, isImportant), http.StatusBadRequest)
		return
	}

	// Determine the importance value
	importance := "normal"
	if isImportant == "true" {
		importance = "high"
	}

	// Update the task importance
	if err := h.Client.UpdateTaskImportance(session.AccessToken, listID, taskID, importance); err != nil {
		http.Error(w, "Error updating task importance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task importance updated successfully"))
}

// GetTaskDetailsHandler handles retrieving details for a specific task
func (h *Handler) GetTaskDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the query parameters
	listID := r.URL.Query().Get("listId")
	taskID := r.URL.Query().Get("taskId")

	if listID == "" || taskID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Get the task details from Microsoft API
	task, err := h.Client.GetTaskDetails(session.AccessToken, listID, taskID)
	if err != nil {
		http.Error(w, "Error fetching task details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response object
	response := map[string]interface{}{
		"id":         task.ID,
		"title":      task.Title,
		"status":     task.Status,
		"importance": task.Importance,
		"categories": task.Categories,
	}

	// Add due date if it exists
	if task.DueDateTime != nil {
		// Store the raw datetime for form handling
		response["dueDateTimeRaw"] = task.DueDateTime.DateTime

		// Parse the due date for display
		t, err := time.Parse(time.RFC3339, task.DueDateTime.DateTime)
		if err == nil {
			// Format as a more readable date
			response["dueDateTime"] = t.Format("Jan 2, 2006")
		} else {
			response["dueDateTime"] = task.DueDateTime.DateTime
		}
	}

	// Convert to JSON and send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateTaskDetailsHandler handles updating a task's details
func (h *Handler) UpdateTaskDetailsHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract form values
	listID := r.FormValue("listId")
	taskID := r.FormValue("taskId")
	title := r.FormValue("title")
	status := r.FormValue("status")
	importance := r.FormValue("importance")
	dueDate := r.FormValue("dueDate")
	categoriesJson := r.FormValue("categories")

	if listID == "" || taskID == "" || title == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Parse categories JSON
	var categories []string
	if categoriesJson != "" {
		if err := json.Unmarshal([]byte(categoriesJson), &categories); err != nil {
			http.Error(w, "Invalid categories format: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Update the task
	if err := h.Client.UpdateTaskDetails(
		session.AccessToken,
		listID,
		taskID,
		title,
		status,
		importance,
		dueDate,
		categories,
	); err != nil {
		http.Error(w, "Error updating task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task updated successfully"))
}

// CreateTaskHandler handles creating a new task
func (h *Handler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract form values
	listID := r.FormValue("listId")
	title := r.FormValue("title")

	if listID == "" || title == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Create the task
	taskID, err := h.Client.CreateTask(session.AccessToken, listID, title)
	if err != nil {
		http.Error(w, "Error creating task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := map[string]string{
		"taskId": taskID,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteTaskHandler handles deleting a task
func (h *Handler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the session ID from the cookie
	sessionID, err := auth.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the session
	session, ok := h.SessionManager.GetSession(sessionID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the session if needed
	if err := h.SessionManager.RefreshSessionIfNeeded(sessionID); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract form values
	listID := r.FormValue("listId")
	taskID := r.FormValue("taskId")

	if listID == "" || taskID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Delete the task
	if err := h.Client.DeleteTask(session.AccessToken, listID, taskID); err != nil {
		http.Error(w, "Error deleting task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task deleted successfully"))
}

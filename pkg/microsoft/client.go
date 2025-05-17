// Package microsoft provides a client for Microsoft Graph API
package microsoft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coseguera/kanban-to-do/internal/models"
)

// Config contains the configuration for the Microsoft client
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	Scope        string
	GraphURL     string
}

// Client is a client for Microsoft Graph API
type Client struct {
	config Config
}

// NewClient creates a new Microsoft client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
	}
}

// GetAuthURL returns the authorization URL
func (c *Client) GetAuthURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&response_mode=query&state=%s",
		c.config.AuthURL, c.config.ClientID, url.QueryEscape(c.config.RedirectURI), url.QueryEscape(c.config.Scope), state)
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (c *Client) ExchangeCodeForToken(code string) (*models.TokenResponse, error) {
	tokenData := url.Values{}
	tokenData.Set("client_id", c.config.ClientID)
	tokenData.Set("client_secret", c.config.ClientSecret)
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", c.config.RedirectURI)
	tokenData.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", c.config.TokenURL, strings.NewReader(tokenData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating token request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code for token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading token response: %w", err)
	}

	var tokenResp models.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an access token
func (c *Client) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	tokenData := url.Values{}
	tokenData.Set("client_id", c.config.ClientID)
	tokenData.Set("client_secret", c.config.ClientSecret)
	tokenData.Set("refresh_token", refreshToken)
	tokenData.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", c.config.TokenURL, strings.NewReader(tokenData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating refresh token request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading refresh token response: %w", err)
	}

	var tokenResp models.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("error parsing refresh token response: %w", err)
	}

	return &tokenResp, nil
}

// GetTodoLists gets the user's to-do lists
func (c *Client) GetTodoLists(accessToken string) (*models.TodoListResponse, error) {
	req, err := http.NewRequest("GET", c.config.GraphURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	var listResp models.TodoListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("error parsing API response: %w", err)
	}

	return &listResp, nil
}

// GetListDetails gets details of a specific to-do list
func (c *Client) GetListDetails(accessToken string, listID string) (*models.TodoList, error) {
	url := fmt.Sprintf("%s/%s", c.config.GraphURL, listID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	var list models.TodoList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("error parsing API response: %w", err)
	}

	return &list, nil
}

// GetListTasks gets the tasks for a specific to-do list
func (c *Client) GetListTasks(accessToken string, listID string) (*models.TaskResponse, error) {
	url := fmt.Sprintf("%s/%s/tasks", c.config.GraphURL, listID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	var taskResp models.TaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return nil, fmt.Errorf("error parsing API response: %w", err)
	}

	return &taskResp, nil
}

// UpdateTaskStatus updates a task's status and categories
func (c *Client) UpdateTaskStatus(accessToken string, listID string, taskID string, status string, categories []string) error {
	url := fmt.Sprintf("%s/%s/tasks/%s", c.config.GraphURL, listID, taskID)

	// Prepare the update payload
	payload := map[string]interface{}{
		"status":     status,
		"categories": categories,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling task update: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating API request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error updating task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	return nil
}

// UpdateTaskImportance updates a task's importance
func (c *Client) UpdateTaskImportance(accessToken string, listID string, taskID string, importance string) error {
	url := fmt.Sprintf("%s/%s/tasks/%s", c.config.GraphURL, listID, taskID)

	// Prepare the update payload
	payload := map[string]interface{}{
		"importance": importance,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling task update: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating API request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error updating task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetTaskDetails retrieves details for a specific task
func (c *Client) GetTaskDetails(accessToken string, listID string, taskID string) (*models.Task, error) {
	url := fmt.Sprintf("%s/%s/tasks/%s", c.config.GraphURL, listID, taskID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading API response: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal(body, &task); err != nil {
		return nil, fmt.Errorf("error parsing API response: %w", err)
	}

	return &task, nil
}

// UpdateTaskDetails updates a task's details
func (c *Client) UpdateTaskDetails(accessToken string, listID string, taskID string, title string, status string, importance string, dueDate string, categories []string) error {
	url := fmt.Sprintf("%s/%s/tasks/%s", c.config.GraphURL, listID, taskID)

	// Build the request body
	requestBody := make(map[string]interface{})
	requestBody["title"] = title
	requestBody["importance"] = importance

	// Handle status
	if status == "completed" {
		requestBody["status"] = "completed"
	} else {
		requestBody["status"] = "notStarted"
	}

	// Handle due date
	if dueDate != "" {
		requestBody["dueDateTime"] = map[string]string{
			"dateTime": dueDate + "T00:00:00Z",
			"timeZone": "UTC",
		}
	}

	// Add categories
	requestBody["categories"] = categories

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error creating JSON request: %w", err)
	}

	// Create PATCH request
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	return nil
}

// CreateTask creates a new task in a list
func (c *Client) CreateTask(accessToken string, listID string, title string) (string, error) {
	url := fmt.Sprintf("%s/%s/tasks", c.config.GraphURL, listID)

	// Build the request body with minimal required fields
	requestBody := map[string]interface{}{
		"title": title,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error creating JSON request: %w", err)
	}

	// Create POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading API response: %w", err)
	}

	// Parse the response to get the task ID
	var task models.Task
	if err := json.Unmarshal(body, &task); err != nil {
		return "", fmt.Errorf("error parsing API response: %w", err)
	}

	return task.ID, nil
}

// DeleteTask deletes a task from a list
func (c *Client) DeleteTask(accessToken string, listID string, taskID string) error {
	url := fmt.Sprintf("%s/%s/tasks/%s", c.config.GraphURL, listID, taskID)

	// Create DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling Microsoft Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body))
	}

	return nil
}

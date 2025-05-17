// Package models contains the data structures used in the application
package models

import "time"

// TodoList represents a Microsoft To Do list
type TodoList struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// TodoListResponse represents the response from the Microsoft Graph API
type TodoListResponse struct {
	Value []TodoList `json:"value"`
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// Session stores user session data
type Session struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

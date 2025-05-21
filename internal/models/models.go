// Copyright (c) 2025 Carlos Oseguera (@coseguera)
// This code is licensed under a dual-license model.
// See LICENSE.md for more information.

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

// Task represents a Microsoft To Do task
type Task struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Status          string    `json:"status"`
	Importance      string    `json:"importance"`
	DueDateTime     *DateTime `json:"dueDateTime,omitempty"`
	CreatedDateTime string    `json:"createdDateTime"`
	Categories      []string  `json:"categories,omitempty"`
}

// DateTime represents a date and time in Microsoft Graph API
type DateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

// TaskResponse represents the response from the Microsoft Graph API for tasks
type TaskResponse struct {
	Value []Task `json:"value"`
}

// KanbanColumn represents a column in the Kanban board
type KanbanColumn struct {
	Title string
	Tasks []TaskDisplay
}

// TaskViewModel is used for rendering tasks in the template
type TaskViewModel struct {
	ListID   string
	ListName string
	Columns  []KanbanColumn
}

// TaskDisplay is a simplified version of Task for display
type TaskDisplay struct {
	ID          string
	Title       string
	Status      bool     // true if completed
	Importance  bool     // true if high importance
	DueDateTime string   // formatted date string
	Categories  []string // list of categories
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

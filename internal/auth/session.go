// Package auth handles authentication and session management
package auth

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coseguera/kanban-to-do/internal/models"
	"github.com/coseguera/kanban-to-do/pkg/microsoft"
)

// SessionManager manages user sessions
type SessionManager struct {
	sessions map[string]models.Session
	mu       sync.RWMutex
	client   *microsoft.Client
}

// NewSessionManager creates a new session manager
func NewSessionManager(client *microsoft.Client) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]models.Session),
		client:   client,
	}
}

// CreateSession creates a new session from a token response
func (sm *SessionManager) CreateSession(tokenResp *models.TokenResponse) (string, error) {
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sessions[sessionID] = models.Session{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return sessionID, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (models.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, ok := sm.sessions[sessionID]
	return session, ok
}

// RefreshSessionIfNeeded refreshes a session if the token is expired
func (sm *SessionManager) RefreshSessionIfNeeded(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found")
	}

	// Check if the token is expired
	if time.Now().After(session.ExpiresAt) {
		tokenResp, err := sm.client.RefreshToken(session.RefreshToken)
		if err != nil {
			return err
		}

		// Update the session
		session.AccessToken = tokenResp.AccessToken
		if tokenResp.RefreshToken != "" {
			session.RefreshToken = tokenResp.RefreshToken
		}
		session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		sm.sessions[sessionID] = session
	}

	return nil
}

// DeleteSession deletes a session by ID
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, sessionID)
}

// SetSessionCookie sets a session cookie
func SetSessionCookie(w http.ResponseWriter, sessionID string) {
	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   86400, // 1 day
	}
	http.SetCookie(w, &cookie)
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	expiredCookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	}
	http.SetCookie(w, &expiredCookie)
}

// GetSessionFromRequest gets the session ID from the request cookie
func GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

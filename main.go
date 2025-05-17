package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Configuration for Microsoft OAuth
var (
	clientID     = os.Getenv("MS_CLIENT_ID")     // Your Azure/Entra App ID
	clientSecret = os.Getenv("MS_CLIENT_SECRET") // Your client secret
	redirectURI  = "https://localhost:8443/auth/callback"
	authURL      = "https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize"
	tokenURL     = "https://login.microsoftonline.com/consumers/oauth2/v2.0/token"
	scope        = "offline_access User.Read Tasks.ReadWrite"
	graphURL     = "https://graph.microsoft.com/v1.0/me/todo/lists"
)

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

var sessions = make(map[string]Session)

func main() {
	// Check if environment variables are set
	if clientID == "" || clientSecret == "" {
		log.Fatal("Please set MS_CLIENT_ID and MS_CLIENT_SECRET environment variables")
	}

	// Create templates directory if it doesn't exist
	if err := os.MkdirAll("templates", 0755); err != nil {
		log.Fatalf("Failed to create templates directory: %v", err)
	}

	// Create certs directory if it doesn't exist
	if err := os.MkdirAll("certs", 0755); err != nil {
		log.Fatalf("Failed to create certs directory: %v", err)
	}

	// Create template files
	createTemplates()

	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth/callback", callbackHandler)
	http.HandleFunc("/todoLists", todoListsHandler)
	http.HandleFunc("/logout", logoutHandler)

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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate a state parameter to prevent CSRF
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create the authorization URL
	authRequestURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&response_mode=query&state=%s",
		authURL, clientID, url.QueryEscape(redirectURI), url.QueryEscape(scope), state)

	http.Redirect(w, r, authRequestURL, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in callback", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenData := url.Values{}
	tokenData.Set("client_id", clientID)
	tokenData.Set("client_secret", clientSecret)
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", redirectURI)
	tokenData.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(tokenData.Encode()))
	if err != nil {
		http.Error(w, "Error creating token request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error exchanging code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading token response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		http.Error(w, "Error parsing token response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a session ID
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Store the token in the session
	sessions[sessionID] = Session{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	// Set a cookie with the session ID
	cookie := http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   86400, // 1 day
	}
	http.SetCookie(w, &cookie)

	// Redirect to the to-do lists page
	http.Redirect(w, r, "/todoLists", http.StatusFound)
}

func todoListsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Get the session
	session, ok := sessions[cookie.Value]
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Check if the token is expired
	if time.Now().After(session.ExpiresAt) {
		// Refresh the token
		if err := refreshToken(cookie.Value, &session); err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	// Call the Microsoft Graph API
	req, err := http.NewRequest("GET", graphURL, nil)
	if err != nil {
		http.Error(w, "Error creating API request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("Authorization", "Bearer "+session.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error calling Microsoft Graph API: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Microsoft Graph API returned error: %s - %s", resp.Status, string(body)), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading API response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var listResp TodoListResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		http.Error(w, "Error parsing API response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	tmpl, err := template.ParseFiles("templates/todoLists.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, listResp)
}

func refreshToken(sessionID string, session *Session) error {
	tokenData := url.Values{}
	tokenData.Set("client_id", clientID)
	tokenData.Set("client_secret", clientSecret)
	tokenData.Set("refresh_token", session.RefreshToken)
	tokenData.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(tokenData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	// Update the session
	session.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		session.RefreshToken = tokenResp.RefreshToken
	}
	session.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	sessions[sessionID] = *session

	return nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Delete the session cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		delete(sessions, cookie.Value)
	}

	// Clear the cookie
	expiredCookie := http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	}
	http.SetCookie(w, &expiredCookie)

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}

func createTemplates() {
	// Create home.html
	homeHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Microsoft To Do Lists</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; line-height: 1.6; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; border-radius: 5px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #0078d4; }
        .login-button { 
            background-color: #0078d4; 
            color: white; 
            padding: 10px 20px; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
            font-size: 16px; 
        }
        .login-button:hover { background-color: #005a9e; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Microsoft To Do Lists</h1>
        <p>Connect to your Microsoft account to see your To Do lists.</p>
        <a href="/login"><button class="login-button">Sign in with Microsoft</button></a>
    </div>
</body>
</html>`

	// Create todoLists.html
	todoListsHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Your To Do Lists</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; line-height: 1.6; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; border-radius: 5px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #0078d4; }
        ul { list-style-type: none; padding: 0; }
        li { 
            margin: 10px 0;
            padding: 15px;
            border-radius: 5px;
            background-color: #f5f5f5;
            border-left: 5px solid #0078d4;
        }
        .list-name { font-weight: bold; font-size: 18px; }
        .list-id { color: #666; font-size: 14px; }
        .logout { 
            background-color: #d9534f;
            color: white;
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-top: 20px;
            display: inline-block;
            text-decoration: none;
        }
        .logout:hover { background-color: #c9302c; }
        .no-lists { 
            padding: 20px;
            background-color: #f9f9f9;
            border-radius: 5px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Your To Do Lists</h1>
        {{if .Value}}
            <ul>
                {{range .Value}}
                    <li>
                        <div class="list-name">{{.DisplayName}}</div>
                        <div class="list-id">ID: {{.ID}}</div>
                    </li>
                {{end}}
            </ul>
        {{else}}
            <div class="no-lists">
                <p>No to-do lists found. Create some in your Microsoft To Do app!</p>
            </div>
        {{end}}
        <a href="/logout" class="logout">Logout</a>
    </div>
</body>
</html>`

	// Write the templates to files
	if err := os.WriteFile("templates/home.html", []byte(homeHTML), 0644); err != nil {
		log.Fatalf("Failed to create home.html template: %v", err)
	}

	if err := os.WriteFile("templates/todoLists.html", []byte(todoListsHTML), 0644); err != nil {
		log.Fatalf("Failed to create todoLists.html template: %v", err)
	}
}

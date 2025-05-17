// Package templates handles the HTML templates used by the application
package templates

import (
	"html/template"
	"os"
	"path/filepath"
)

var (
	// Templates are the parsed HTML templates
	Templates map[string]*template.Template
)

// LoadTemplates loads and parses all HTML templates
func LoadTemplates(templatesDir string) error {
	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	// Create the templates
	if err := createTemplates(templatesDir); err != nil {
		return err
	}

	// Parse the templates
	Templates = make(map[string]*template.Template)

	homeTmpl, err := template.ParseFiles(filepath.Join(templatesDir, "home.html"))
	if err != nil {
		return err
	}
	Templates["home"] = homeTmpl

	todoListsTmpl, err := template.ParseFiles(filepath.Join(templatesDir, "todoLists.html"))
	if err != nil {
		return err
	}
	Templates["todoLists"] = todoListsTmpl

	return nil
}

func createTemplates(templatesDir string) error {
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
	if err := os.WriteFile(filepath.Join(templatesDir, "home.html"), []byte(homeHTML), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(templatesDir, "todoLists.html"), []byte(todoListsHTML), 0644); err != nil {
		return err
	}

	return nil
}

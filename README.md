# Kanban To-Do

A Go web server that integrates with Microsoft To-Do using the Microsoft Graph API to display a user's to-do lists.

## Project Structure

The project follows a standard Go project layout:

```
kanban-to-do/
├── cmd/
│   └── server/        # Application entry point
├── internal/
│   ├── auth/          # Authentication and session management
│   ├── handlers/      # HTTP handlers
│   ├── models/        # Data models
│   └── templates/     # HTML templates
├── pkg/
│   └── microsoft/     # Microsoft Graph API client
├── certs/             # SSL certificates (not checked into git)
└── templates/         # Compiled HTML templates
```

## Features

- HTTPS-only web server
- Authentication with Microsoft personal accounts
- Display of user's to-do lists from Microsoft To-Do
- Session management
- Automatic token refresh

## Prerequisites

- Go 1.24+
- Microsoft/Entra App Registration with client ID and client secret
- OpenSSL (for generating self-signed certificates)

## Setup

1. Register an application in the Microsoft Entra ID (Azure AD) portal:
   - Navigate to [Azure Portal](https://portal.azure.com)
   - Go to Entra ID / App Registrations
   - Register a new application
   - Set a redirect URI: `https://localhost:8443/auth/callback`
   - Ensure you have the necessary API permissions: `Tasks.ReadWrite` and `User.Read`
   - Create a client secret and save it securely

2. Generate self-signed certificates for HTTPS:
   ```bash
   chmod +x generate_cert.sh
   ./generate_cert.sh
   ```

3. Set environment variables for your Microsoft application (do not commit these values to git):
   ```bash
   export MS_CLIENT_ID="your-client-id"
   export MS_CLIENT_SECRET="your-client-secret"
   ```
   
   Alternatively, you can create a `.env` file (this file should be listed in .gitignore and never committed):
   ```
   MS_CLIENT_ID=your-client-id
   MS_CLIENT_SECRET=your-client-secret
   ```

## Running the Application

1. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

2. Open your browser and navigate to:
   ```
   https://localhost:8443
   ```

3. Click "Sign in with Microsoft" and follow the authentication flow

## Notes

- This application uses a self-signed certificate for HTTPS, which will generate browser warnings in a development environment
- For production, replace the self-signed certificate with a proper one from a certificate authority
- Token refresh is handled automatically when tokens expire
- User sessions are stored in memory and will be lost when the server restarts

## License

This project is available under a dual license:
- Free for non-commercial use under the MIT License
- Requires a paid commercial license for any for-profit use

See the [LICENSE.md](LICENSE.md) file for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

## Security

For security issues, please refer to [SECURITY.md](SECURITY.md).

## Security Considerations

- **Never commit secrets or certificate files to the repository**
- Before running the application for the first time, generate new certificates using the provided script
- Store environment variables securely and never include them in the repository
- The self-signed certificates generated by the script are intended for development only
- For production use, obtain proper certificates from a trusted certificate authority
- To prepare the repository for public sharing, run the provided script: `./prepare_for_public.sh`

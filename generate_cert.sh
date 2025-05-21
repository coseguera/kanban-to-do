#!/bin/bash
# Copyright (c) 2025 Carlos Oseguera (@coseguera)
# This code is licensed under a dual-license model.
# See LICENSE.md for more information.

# Script to generate self-signed certificates for local HTTPS development

# Create the certs directory if it doesn't exist
mkdir -p certs

# Generate a self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj '/CN=localhost'

echo "Self-signed certificate generated successfully!"
echo "Certificate: certs/server.crt"
echo "Private key: certs/server.key"

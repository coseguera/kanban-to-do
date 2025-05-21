#!/bin/bash
# Copyright (c) 2025 Carlos Oseguera (@coseguera)
# This code is licensed under a dual-license model.
# See LICENSE.md for more information.

# Script to prepare the repository for public release by removing sensitive files

echo "Preparing the repository for public release..."

# Remove certificate files
if [ -d "certs" ]; then
    echo "Removing certificate files..."
    rm -rf certs
    echo "Certificate files removed."
fi

# Check for any potential .env files
ENV_FILES=$(find . -name "*.env" -o -name ".env*")
if [ -n "$ENV_FILES" ]; then
    echo "WARNING: The following environment files were found and should be removed:"
    echo "$ENV_FILES"
    echo "To remove them, run: rm $ENV_FILES"
fi

# Create certs directory but keep it empty
mkdir -p certs
touch certs/.gitkeep

echo "Repository is now ready for public release."
echo "Remember to run this script again if you've made changes that might introduce sensitive information."
echo ""
echo "IMPORTANT: Before pushing, verify you have not committed any credentials or secrets!"

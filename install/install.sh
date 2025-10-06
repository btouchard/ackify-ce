#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later

# Ackify Community Edition (CE) Installation Script
# Quick setup for Docker deployment

set -e

echo "🔐 Ackify Community Edition (CE) Installation"
echo "========================="

# Create installation directory
INSTALL_DIR="ackify-ce"
if [ -d "$INSTALL_DIR" ]; then
    echo "❌ Directory $INSTALL_DIR already exists. Please remove it first."
    exit 1
fi

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo "📦 Downloading configuration files..."

# Download docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify-ce/main/install/docker-compose.yml -o docker-compose.yml

# Download .env.example
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify-ce/main/install/.env.example -o .env.example

echo "🔧 Setting up environment..."

# Copy .env.example to .env
cp .env.example .env

# Generate secure secrets
echo "🔑 Generating secure secrets..."
COOKIE_SECRET=$(openssl rand -base64 32)
ED25519_KEY=$(openssl rand 64 | base64 -w 0)

# Replace placeholders in .env (using # as delimiter to avoid issues with / and + in base64)
sed -i "s#your_base64_encoded_secret_key#$COOKIE_SECRET#" .env
sed -i "s#your_base64_encoded_ed25519_private_key#$ED25519_KEY#" .env

# Generate random password for PostgreSQL
DB_PASSWORD=$(openssl rand -base64 24)
sed -i "s#your_secure_password#$DB_PASSWORD#" .env

echo "✅ Installation completed!"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file with your OAuth2 configuration:"
echo "   - Set APP_DNS to your domain"
echo "   - Configure ACKIFY_OAUTH_CLIENT_ID and ACKIFY_OAUTH_CLIENT_SECRET"
echo "   - Optionally set ACKIFY_OAUTH_ALLOWED_DOMAIN for user restriction"
echo ""
echo "2. Start Ackify:"
echo "   docker compose up -d"
echo ""
echo "3. Check health:"
echo "   curl http://localhost:8080/health   # alias: /health"
echo ""
echo "📁 Installation directory: $(pwd)"

#!/bin/bash

# Ackify Installation Script
# Quick setup for Docker deployment

set -e

echo "🔐 Ackify Installation"
echo "========================="

# Create installation directory
INSTALL_DIR="ackify-install"
if [ -d "$INSTALL_DIR" ]; then
    echo "❌ Directory $INSTALL_DIR already exists. Please remove it first."
    exit 1
fi

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo "📦 Downloading configuration files..."

# Download docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify/main/install/docker-compose.yml -o docker-compose.yml

# Download .env.example
curl -fsSL https://raw.githubusercontent.com/btouchard/ackify/main/install/.env.example -o .env.example

echo "🔧 Setting up environment..."

# Copy .env.example to .env
cp .env.example .env

# Generate secure secrets
echo "🔑 Generating secure secrets..."
COOKIE_SECRET=$(openssl rand -base64 32)
ED25519_KEY=$(openssl genpkey -algorithm Ed25519 | base64 -w 0)

# Replace placeholders in .env
sed -i "s/your_base64_encoded_secret_key/$COOKIE_SECRET/" .env
sed -i "s/your_base64_encoded_ed25519_private_key/$ED25519_KEY/" .env

# Generate random password for PostgreSQL
DB_PASSWORD=$(openssl rand -base64 24)
sed -i "s/your_secure_password/$DB_PASSWORD/" .env

echo "✅ Installation completed!"
echo ""
echo "📋 Next steps:"
echo "1. Edit .env file with your OAuth2 configuration:"
echo "   - Set APP_DNS to your domain"
echo "   - Configure OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET"
echo "   - Optionally set OAUTH_ALLOWED_DOMAIN for user restriction"
echo ""
echo "2. Start Ackify:"
echo "   docker compose up -d"
echo ""
echo "3. Check health:"
echo "   curl http://localhost:8080/healthz"
echo ""
echo "📁 Installation directory: $(pwd)"
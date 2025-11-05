#!/bin/bash
# SPDX-License-Identifier: AGPL-3.0-or-later

# Ackify Community Edition (CE) Installation Script
# Interactive setup for Docker deployment

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

prompt_input() {
    local prompt="$1"
    local default="$2"
    local response

    if [ -n "$default" ]; then
        read -p "$(echo -e ${BLUE}"$prompt [${default}]: "${NC})" response
        echo "${response:-$default}"
    else
        read -p "$(echo -e ${BLUE}"$prompt: "${NC})" response
        echo "$response"
    fi
}

prompt_yes_no() {
    local prompt="$1"
    local default="$2"
    local response

    if [ "$default" = "y" ]; then
        read -p "$(echo -e ${BLUE}"$prompt [Y/n]: "${NC})" response
        response="${response:-y}"
    else
        read -p "$(echo -e ${BLUE}"$prompt [y/N]: "${NC})" response
        response="${response:-n}"
    fi

    [[ "$response" =~ ^[Yy]$ ]]
}

prompt_password() {
    local prompt="$1"
    local response

    read -sp "$(echo -e ${BLUE}"$prompt: "${NC})" response
    echo "" >&2
    echo "$response"
}

# Main installation
clear
print_header "🔐 Ackify Community Edition (CE) - Interactive Installation"

echo ""
print_info "This script will guide you through the configuration of Ackify CE."
print_info "At least ONE authentication method (OAuth or MagicLink) must be enabled."
echo ""

# Create installation directory
INSTALL_DIR="ackify-ce"
if [ -d "$INSTALL_DIR" ]; then
    print_error "Directory $INSTALL_DIR already exists."
    if prompt_yes_no "Remove existing directory and continue?" "n"; then
        rm -rf "$INSTALL_DIR"
        print_success "Existing directory removed"
    else
        print_error "Installation cancelled"
        exit 1
    fi
fi

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

print_success "Installation directory created: $(pwd)"
echo ""

# ==========================================
# Basic Configuration
# ==========================================
print_header "🌐 Basic Configuration"
echo ""

APP_BASE_URL=$(prompt_input "Application Base URL (e.g., https://ackify.example.com)" "http://localhost:8080")
APP_ORGANISATION=$(prompt_input "Organization Name" "My Organization")
APP_DNS=$(echo "$APP_BASE_URL" | sed -E 's|https?://||')

# Extract domain for email addresses (remove subdomain if present, remove port)
APP_DOMAIN=$(echo "$APP_DNS" | sed 's/:.*//')
# Count dots to detect subdomain
dot_count=$(echo "$APP_DOMAIN" | tr -cd '.' | wc -c)
# If 2+ dots (subdomain.domain.tld), remove first part to keep domain.tld
if [ "$dot_count" -ge 2 ]; then
    APP_DOMAIN=$(echo "$APP_DOMAIN" | cut -d. -f2-)
fi

print_success "Base configuration set"
echo ""

# ==========================================
# Traefik Configuration (early choice)
# ==========================================
print_header "🔀 Reverse Proxy Configuration (Traefik)"
echo ""
print_info "Ackify can be configured to work with Traefik for automatic HTTPS."
print_info "If you choose Traefik, ports will not be exposed (Traefik will handle routing)."
echo ""

ENABLE_TRAEFIK=false
COMPOSE_TEMPLATE="compose.yml"

if prompt_yes_no "Use Traefik reverse proxy?" "n"; then
    ENABLE_TRAEFIK=true
    COMPOSE_TEMPLATE="compose-traefik.yml"

    echo ""
    TRAEFIK_NETWORK=$(prompt_input "Docker network where Traefik is running" "traefik")
    TRAEFIK_CERTRESOLVER=$(prompt_input "Traefik TLS cert resolver name" "letsencrypt")

    # Extract app name from DNS
    APP_NAME=$(echo "$APP_DNS" | sed 's/\..*//')

    print_success "Traefik configuration completed"
    echo ""
    print_info "Traefik network: ${TRAEFIK_NETWORK}"
    print_info "TLS cert resolver: ${TRAEFIK_CERTRESOLVER}"
    print_info "App name: ${APP_NAME}"
else
    print_info "Standard configuration selected (port 8080 will be exposed)"
fi
echo ""

# Download configuration files
print_header "📦 Downloading Configuration Files"
echo ""

curl -fsSL "https://raw.githubusercontent.com/btouchard/ackify-ce/main/install/${COMPOSE_TEMPLATE}" -o compose.yml
print_success "compose.yml downloaded (${COMPOSE_TEMPLATE})"

curl -fsSL https://raw.githubusercontent.com/btouchard/ackify-ce/main/install/.env.example -o .env.example
print_success ".env.example downloaded"
echo ""

# ==========================================
# OAuth Configuration
# ==========================================
print_header "🔑 OAuth2 Authentication Configuration"
echo ""
print_info "OAuth allows users to sign in with Google, GitHub, GitLab, or custom provider."
echo ""

ENABLE_OAUTH=false
if prompt_yes_no "Enable OAuth2 authentication?" "y"; then
    ENABLE_OAUTH=true

    echo ""
    print_info "Available providers: google, github, gitlab, custom"
    OAUTH_PROVIDER=$(prompt_input "OAuth Provider (google/github/gitlab/custom)" "custom")

    OAUTH_CLIENT_ID=$(prompt_input "OAuth Client ID")
    OAUTH_CLIENT_SECRET=$(prompt_input "OAuth Client Secret")

    echo ""
    if prompt_yes_no "Restrict to specific email domain? (e.g., @company.com)" "n"; then
        OAUTH_ALLOWED_DOMAIN=$(prompt_input "Allowed email domain (e.g., @company.com)" "@example.com")
    else
        OAUTH_ALLOWED_DOMAIN=""
    fi

    # Custom provider configuration
    if [ "$OAUTH_PROVIDER" = "custom" ]; then
        echo ""
        print_info "Custom OAuth provider configuration"
        OAUTH_AUTH_URL=$(prompt_input "Authorization URL")
        OAUTH_TOKEN_URL=$(prompt_input "Token URL")
        OAUTH_USERINFO_URL=$(prompt_input "User Info URL")
        OAUTH_SCOPES=$(prompt_input "OAuth Scopes" "openid,email,profile")
    fi

    # GitLab self-hosted
    if [ "$OAUTH_PROVIDER" = "gitlab" ]; then
        echo ""
        if prompt_yes_no "Using self-hosted GitLab?" "n"; then
            OAUTH_GITLAB_URL=$(prompt_input "GitLab instance URL" "https://git.company.com")
        fi
    fi

    echo ""
    if prompt_yes_no "Enable auto-login (automatically log in if OAuth session exists)?" "n"; then
        OAUTH_AUTO_LOGIN="true"
    else
        OAUTH_AUTO_LOGIN="false"
    fi

    print_success "OAuth configuration completed"
else
    print_warning "OAuth authentication disabled"
fi
echo ""

# ==========================================
# SMTP Configuration
# ==========================================
print_header "📧 SMTP Configuration (Email Service)"
echo ""
print_info "SMTP is used for:"
print_info "  - Sending signature reminders to expected signers"
print_info "  - MagicLink authentication (passwordless email login)"
echo ""

ENABLE_SMTP=false
if prompt_yes_no "Enable SMTP for email notifications and MagicLink?" "y"; then
    ENABLE_SMTP=true

    echo ""
    MAIL_HOST=$(prompt_input "SMTP Host (e.g., smtp.gmail.com)")
    MAIL_PORT=$(prompt_input "SMTP Port" "587")
    MAIL_USERNAME=$(prompt_input "SMTP Username")
    MAIL_PASSWORD=$(prompt_password "SMTP Password")

    echo ""
    MAIL_FROM=$(prompt_input "From Email Address" "noreply@${APP_DOMAIN}")
    MAIL_FROM_NAME=$(prompt_input "From Name" "$APP_ORGANISATION")

    echo ""
    if prompt_yes_no "Use TLS?" "y"; then
        MAIL_TLS="true"
    else
        MAIL_TLS="false"
    fi

    if prompt_yes_no "Use STARTTLS?" "y"; then
        MAIL_STARTTLS="true"
    else
        MAIL_STARTTLS="false"
    fi

    print_success "SMTP configuration completed"
    echo ""

    # MagicLink configuration
    print_header "🔗 MagicLink Authentication"
    echo ""
    print_info "MagicLink provides passwordless authentication via email."
    print_info "Since SMTP is configured, MagicLink will be enabled by default."
    echo ""

    ENABLE_MAGICLINK=true
    if prompt_yes_no "Disable MagicLink authentication?" "n"; then
        ENABLE_MAGICLINK=false
        print_warning "MagicLink authentication will be disabled"
    else
        print_success "MagicLink authentication enabled"
    fi
else
    print_warning "SMTP disabled - Email reminders and MagicLink will not be available"
    ENABLE_MAGICLINK=false
fi
echo ""

# ==========================================
# Authentication Method Validation
# ==========================================
if [ "$ENABLE_OAUTH" = false ] && [ "$ENABLE_MAGICLINK" = false ]; then
    print_error "At least ONE authentication method must be enabled!"
    print_error "Please enable OAuth or configure SMTP for MagicLink."
    exit 1
fi

# ==========================================
# Admin Configuration
# ==========================================
print_header "👤 Admin Configuration"
echo ""
print_info "Admin users have access to document management and reminder features."
print_info "At least ONE admin email address is required."
echo ""

ADMIN_EMAILS=""
while [ -z "$ADMIN_EMAILS" ]; do
    ADMIN_EMAILS=$(prompt_input "Admin email addresses (comma-separated)" "admin@${APP_DOMAIN}")

    if [ -z "$ADMIN_EMAILS" ]; then
        print_error "At least one admin email is required!"
        echo ""
    fi
done

print_success "Admin users configured: $ADMIN_EMAILS"
echo ""

# ==========================================
# Generate Secrets
# ==========================================
print_header "🔑 Generating Secure Secrets"
echo ""

COOKIE_SECRET=$(openssl rand -base64 32)
print_success "Cookie secret generated"

ED25519_KEY=$(openssl rand 64 | base64 -w 0)
print_success "Ed25519 private key generated"

DB_PASSWORD=$(openssl rand -hex 24)
print_success "Database password generated"
echo ""

# ==========================================
# Create .env file
# ==========================================
print_header "📝 Creating Configuration File"
echo ""

cat > .env <<EOF
# ==========================================
# Ackify Community Edition Configuration
# Generated on $(date)
# ==========================================

# ==========================================
# Application Configuration
# ==========================================
ACKIFY_BASE_URL=${APP_BASE_URL}
ACKIFY_ORGANISATION=${APP_ORGANISATION}

# ==========================================
# Database Configuration
# ==========================================
POSTGRES_USER=ackifyr
POSTGRES_PASSWORD=${DB_PASSWORD}
POSTGRES_DB=ackify
ACKIFY_DB_DSN=postgres://ackifyr:${DB_PASSWORD}@postgres:5432/ackify?sslmode=disable

# ==========================================
# Security Configuration (Auto-generated)
# ==========================================
ACKIFY_OAUTH_COOKIE_SECRET=${COOKIE_SECRET}
ACKIFY_ED25519_PRIVATE_KEY=${ED25519_KEY}

# ==========================================
# Server Configuration
# ==========================================
ACKIFY_LISTEN_ADDR=:8080
ACKIFY_LOG_LEVEL=info

EOF

# OAuth configuration
if [ "$ENABLE_OAUTH" = true ]; then
    cat >> .env <<EOF
# ==========================================
# OAuth2 Configuration
# ==========================================
ACKIFY_OAUTH_PROVIDER=${OAUTH_PROVIDER}
ACKIFY_OAUTH_CLIENT_ID=${OAUTH_CLIENT_ID}
ACKIFY_OAUTH_CLIENT_SECRET=${OAUTH_CLIENT_SECRET}
EOF

    if [ -n "$OAUTH_ALLOWED_DOMAIN" ]; then
        echo "ACKIFY_OAUTH_ALLOWED_DOMAIN=${OAUTH_ALLOWED_DOMAIN}" >> .env
    fi

    if [ "$OAUTH_PROVIDER" = "custom" ]; then
        cat >> .env <<EOF
ACKIFY_OAUTH_AUTH_URL=${OAUTH_AUTH_URL}
ACKIFY_OAUTH_TOKEN_URL=${OAUTH_TOKEN_URL}
ACKIFY_OAUTH_USERINFO_URL=${OAUTH_USERINFO_URL}
ACKIFY_OAUTH_SCOPES=${OAUTH_SCOPES}
EOF
    fi

    if [ -n "$OAUTH_GITLAB_URL" ]; then
        echo "ACKIFY_OAUTH_GITLAB_URL=${OAUTH_GITLAB_URL}" >> .env
    fi

    echo "ACKIFY_OAUTH_AUTO_LOGIN=${OAUTH_AUTO_LOGIN}" >> .env
    echo "" >> .env
fi

# SMTP configuration
if [ "$ENABLE_SMTP" = true ]; then
    cat >> .env <<EOF
# ==========================================
# SMTP Configuration (Email Service)
# ==========================================
ACKIFY_MAIL_HOST=${MAIL_HOST}
ACKIFY_MAIL_PORT=${MAIL_PORT}
ACKIFY_MAIL_USERNAME=${MAIL_USERNAME}
ACKIFY_MAIL_PASSWORD=${MAIL_PASSWORD}
ACKIFY_MAIL_FROM=${MAIL_FROM}
ACKIFY_MAIL_FROM_NAME=${MAIL_FROM_NAME}
ACKIFY_MAIL_TLS=${MAIL_TLS}
ACKIFY_MAIL_STARTTLS=${MAIL_STARTTLS}
ACKIFY_MAIL_TIMEOUT=10s
ACKIFY_MAIL_TEMPLATE_DIR=templates
ACKIFY_MAIL_DEFAULT_LOCALE=en

EOF
fi

# MagicLink configuration
if [ "$ENABLE_MAGICLINK" = "false" ] && [ "$ENABLE_SMTP" = true ]; then
    echo "# Disable MagicLink even though SMTP is configured" >> .env
    echo "ACKIFY_AUTH_MAGICLINK_ENABLED=false" >> .env
    echo "" >> .env
fi

# Admin configuration
if [ -n "$ADMIN_EMAILS" ]; then
    cat >> .env <<EOF
# ==========================================
# Admin Configuration
# ==========================================
ACKIFY_ADMIN_EMAILS=${ADMIN_EMAILS}

EOF
fi

# Traefik configuration
if [ "$ENABLE_TRAEFIK" = true ]; then
    cat >> .env <<EOF
# ==========================================
# Traefik Configuration
# ==========================================
TRAEFIK_NETWORK=${TRAEFIK_NETWORK}
TRAEFIK_CERTRESOLVER=${TRAEFIK_CERTRESOLVER}
APP_NAME=${APP_NAME}
APP_DNS=${APP_DNS}

EOF
fi

print_success ".env file created successfully"
echo ""

# ==========================================
# Installation Summary
# ==========================================
print_header "📊 Installation Summary"
echo ""

print_info "Base URL: ${APP_BASE_URL}"
print_info "Organization: ${APP_ORGANISATION}"
echo ""

print_info "Authentication Methods:"
if [ "$ENABLE_OAUTH" = true ]; then
    print_success "  ✓ OAuth2 ($OAUTH_PROVIDER)"
else
    print_warning "  ✗ OAuth2 (disabled)"
fi

if [ "$ENABLE_MAGICLINK" = "true" ]; then
    print_success "  ✓ MagicLink (passwordless email)"
else
    print_warning "  ✗ MagicLink (disabled)"
fi
echo ""

if [ "$ENABLE_SMTP" = true ]; then
    print_success "Email Service: Enabled (${MAIL_HOST})"
else
    print_warning "Email Service: Disabled"
fi
echo ""

print_success "Admin Users: ${ADMIN_EMAILS}"
echo ""

if [ "$ENABLE_TRAEFIK" = true ]; then
    print_success "Reverse Proxy: Traefik (network: ${TRAEFIK_NETWORK})"
    print_info "TLS Certificate: ${TRAEFIK_CERTRESOLVER}"
else
    print_info "Reverse Proxy: None (direct port 8080 exposure)"
fi
echo ""

# ==========================================
# Next Steps
# ==========================================
print_header "🚀 Next Steps"
echo ""

print_info "1. Review configuration:"
echo "   cat .env"
echo ""

print_info "2. Start Ackify:"
echo "   docker compose up -d"
echo ""

print_info "3. Check logs:"
echo "   docker compose logs -f ackify-ce"
echo ""

print_info "4. Access application:"
echo "   ${APP_BASE_URL}"
echo ""

print_info "5. Check health:"
echo "   curl ${APP_BASE_URL}/health"
echo ""

print_header "✅ Installation Complete!"
echo ""
print_success "Installation directory: $(pwd)"
print_success "Configuration file: $(pwd)/.env"
echo ""

if [ "$ENABLE_OAUTH" = false ] && [ "$ENABLE_MAGICLINK" = true ]; then
    print_warning "Note: Only MagicLink is enabled. Users will need to receive an email to sign in."
    echo ""
fi

print_info "Ready to start Ackify? Run: docker compose up -d"
echo ""

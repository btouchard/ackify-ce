#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building Tailwind CSS...${NC}"

# Check if tailwindcss binary exists, if not download it
if [ ! -f "bin/tailwindcss" ]; then
    echo "Downloading Tailwind CSS CLI v3.4.16..."
    mkdir -p bin
    curl -sL https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.16/tailwindcss-linux-x64 -o bin/tailwindcss
    chmod +x bin/tailwindcss
fi

# Build CSS
mkdir -p ./static

if [ "$1" = "--watch" ]; then
    echo -e "${YELLOW}Watching for changes...${NC}"
    ./bin/tailwindcss -i ./assets/input.css -o ./static/output.css --watch
else
    # Production build with minification
    ./bin/tailwindcss -i ./assets/input.css -o ./static/output.css --minify
    echo -e "${GREEN}✓ CSS built successfully at static/output.css${NC}"
fi

#!/bin/bash

echo "📦 Installing Swag for API Documentation"
echo "========================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    echo "   Visit: https://golang.org/dl/"
    exit 1
fi

echo "✅ Go is installed"

# Install swag
echo "📦 Installing swag..."
go install github.com/swaggo/swag/cmd/swag@latest

if [ $? -ne 0 ]; then
    echo "❌ Failed to install swag"
    exit 1
fi

echo "✅ Swag installed successfully"

# Check if swag is in PATH
if ! command -v swag &> /dev/null; then
    echo "⚠️  Swag installed but not in PATH"
    echo "   Adding $HOME/go/bin to PATH..."
    
    # Add to current session
    export PATH="$PATH:$HOME/go/bin"
    
    # Add to shell profile
    if [[ "$SHELL" == *"zsh"* ]]; then
        echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.zshrc
        echo "✅ Added to ~/.zshrc"
    elif [[ "$SHELL" == *"bash"* ]]; then
        echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
        echo "✅ Added to ~/.bashrc"
    fi
    
    echo "   Please restart your terminal or run: source ~/.zshrc (or ~/.bashrc)"
fi

# Test swag command
if command -v swag &> /dev/null; then
    echo "✅ Swag is now available in PATH"
    swag version
else
    echo "❌ Swag is still not available. Please run:"
    echo "   export PATH=\"\$PATH:\$HOME/go/bin\""
    echo "   Then try again"
fi

echo ""
echo "🎉 Swag installation completed!"
echo "   You can now run: ./scripts/generate_docs.sh"

#!/bin/bash

echo "📚 Generating Swagger Documentation..."
echo "====================================="

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
    
    # Add Go bin to PATH if not already there
    if [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH="$PATH:$HOME/go/bin"
    fi
    
    # Check again after installation
    if ! command -v swag &> /dev/null; then
        echo "❌ Failed to install swag. Please install manually:"
        echo "   go install github.com/swaggo/swag/cmd/swag@latest"
        echo "   Then add $HOME/go/bin to your PATH"
        exit 1
    fi
fi

# Generate swagger docs
echo "Generating API documentation..."
swag init -g main.go -o ./docs

if [ $? -eq 0 ]; then
    echo "✅ Swagger documentation generated successfully!"
    echo ""
    echo "📖 Documentation available at:"
    echo "   http://localhost:8080/swagger/index.html"
    echo "   http://localhost:8080/docs"
else
    echo "❌ Failed to generate documentation"
    exit 1
fi

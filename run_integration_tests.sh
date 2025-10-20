#!/bin/bash

# Integration Test Runner for Instagrano
# This script runs integration tests for the JWT authentication and file upload functionality

echo "🧪 Running Instagrano Integration Tests"
echo "========================================"

# Check if server is running
echo "🔍 Checking if server is running..."
if ! curl -s http://localhost:3007/health > /dev/null; then
    echo "❌ Server is not running!"
    echo ""
    echo "Please start the server first:"
    echo "JWT_SECRET='super-secret-key-for-testing' PORT=3007 go run cmd/api/main.go"
    echo ""
    echo "Then run this script again."
    exit 1
fi

echo "✅ Server is running"
echo ""

# Run the integration tests
echo "🚀 Running integration tests..."
go test -v ./tests/

echo ""
echo "✨ Integration tests completed!"

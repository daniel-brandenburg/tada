#!/bin/bash

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Display coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"


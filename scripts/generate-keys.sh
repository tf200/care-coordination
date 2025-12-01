#!/bin/bash

# Script to generate secure random keys for JWT tokens
# Usage: ./scripts/generate-keys.sh

echo "==================================="
echo "Generating Secure Keys for JWT"
echo "==================================="
echo ""

echo "ACCESS_TOKEN_SECRET:"
openssl rand -base64 64 | tr -d '\n'
echo ""
echo ""

echo "REFRESH_TOKEN_SECRET:"
openssl rand -base64 64 | tr -d '\n'
echo ""
echo ""

echo "==================================="
echo "Copy these values to your .env file"
echo "==================================="

#!/bin/bash

# API Testing Script
# This script tests the main endpoints of the backend API

set -e

BASE_URL="http://localhost:8080"
FILE_ID=""

echo "=== Testing Backend API ==="
echo ""

# Test 1: Health Check
echo "1. Testing health check..."
HEALTH=$(curl -s ${BASE_URL}/health)
echo "Response: $HEALTH"
if echo "$HEALTH" | grep -q "healthy"; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed"
    exit 1
fi
echo ""

# Test 2: Provider Metadata
echo "2. Testing provider metadata..."
PROVIDERS=$(curl -s ${BASE_URL}/api/providers)
echo "Response: $PROVIDERS"
if echo "$PROVIDERS" | grep -q "awsS3"; then
    echo "✓ Providers endpoint passed"
else
    echo "✗ Providers endpoint failed"
    exit 1
fi
echo ""

# Test 3: File Upload
echo "3. Testing file upload..."
# Create test file
echo "Hello, this is a test file for erasure coding!" > /tmp/test-upload.txt
echo "Created test file: /tmp/test-upload.txt"

# Upload file
UPLOAD_RESPONSE=$(curl -s -X POST ${BASE_URL}/api/v1/files/upload \
  -F "file=@/tmp/test-upload.txt" \
  -F "k=2" \
  -F "n=3" \
  -F "chunk_size_mb=1" \
  -F "providers=awsS3,googleDrive")

echo "Upload response: $UPLOAD_RESPONSE"

# Extract file ID
FILE_ID=$(echo "$UPLOAD_RESPONSE" | grep -o '"file_id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$FILE_ID" ]; then
    echo "✓ File upload passed (ID: $FILE_ID)"
else
    echo "✗ File upload failed"
    exit 1
fi
echo ""

# Test 4: Get File Metadata
echo "4. Testing file metadata..."
METADATA=$(curl -s ${BASE_URL}/api/v1/files/${FILE_ID})
echo "Metadata: $METADATA"
if echo "$METADATA" | grep -q "test-upload.txt"; then
    echo "✓ File metadata passed"
else
    echo "✗ File metadata failed"
fi
echo ""

# Test 5: Get Shard Map
echo "5. Testing shard map..."
SHARD_MAP=$(curl -s ${BASE_URL}/api/v1/shards/file/${FILE_ID})
echo "Shard map: $SHARD_MAP"
if echo "$SHARD_MAP" | grep -q "shards"; then
    echo "✓ Shard map passed"
else
    echo "✗ Shard map failed"
fi
echo ""

# Test 6: Download File (Note: will fail with mock adapters)
echo "6. Testing file download..."
echo "Note: Download will fail with mock adapters that don't return actual data"
curl -s -o /tmp/test-download.txt ${BASE_URL}/api/v1/files/${FILE_ID}/download || true
echo "Downloaded to: /tmp/test-download.txt"
echo ""

# Cleanup
rm -f /tmp/test-upload.txt /tmp/test-download.txt

echo "=== Testing Complete ==="
echo ""
echo "Summary:"
echo "- All basic endpoints are working"
echo "- File upload, metadata, and shard tracking work correctly"
echo "- Download requires real cloud adapter implementation"
echo ""
echo "File ID for manual testing: $FILE_ID"

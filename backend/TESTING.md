# Testing Guide

This guide explains how to verify that the implementation works correctly.

## Prerequisites

1. **PostgreSQL Database** running and accessible
2. **Go 1.21+** installed
3. **Environment variables** configured

## Step 1: Database Setup

### Create Database
```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE cs464_project;

# Exit psql
\q
```

### Set Environment Variable
Create a `.env` file in the backend directory:
```bash
cd backend
cat > .env << 'ENVFILE'
DATABASE_URL=postgres://postgres:password@localhost:5432/cs464_project?sslmode=disable
PORT=8080
ENVFILE
```

### Run Migrations
```bash
# Apply database schema
psql -U postgres -d cs464_project -f migrations/001_initial_schema.sql
```

## Step 2: Build and Start Server

```bash
# Build the server
go build -o bin/server ./cmd/server

# Run the server
./bin/server
```

Expected output:
```
Connecting to database...
Database connected successfully
Registered cloud providers: awsS3, googleDrive
Server starting on port 8080
Endpoints available:
  - GET  /health
  - GET  /api/providers
  - POST /api/v1/shards/register
  ...
```

## Step 3: Test Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

## Step 4: Test Provider Metadata

```bash
curl http://localhost:8080/api/providers
```

Expected response:
```json
[
  {
    "providerId": "awsS3",
    "displayName": "AWS S3",
    "status": "connected",
    ...
  },
  {
    "providerId": "googleDrive",
    "displayName": "Google Drive",
    "status": "connected",
    ...
  }
]
```

## Step 5: Test File Upload (End-to-End)

### Create a test file
```bash
echo "Hello, this is a test file for erasure coding!" > test.txt
```

### Upload the file
```bash
curl -X POST http://localhost:8080/api/v1/files/upload \
  -F "file=@test.txt" \
  -F "k=2" \
  -F "n=3" \
  -F "chunk_size_mb=1" \
  -F "providers=awsS3,googleDrive"
```

Expected response:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "original_name": "test.txt",
  "original_size": 47,
  "total_chunks": 1,
  "total_shards": 3,
  "n": 3,
  "k": 2,
  "chunk_size": 1048576,
  "shard_size": 24,
  "status": "UPLOADED",
  "message": "Successfully uploaded 3 shards across 1 chunks",
  "upload_stats": {
    "total_shards": 3,
    "successful_shards": 3,
    "failed_shards": 0,
    "provider_stats": {
      "awsS3": 2,
      "googleDrive": 1
    },
    "duration_ms": 45
  }
}
```

### Get file metadata
```bash
# Use the file_id from upload response
curl http://localhost:8080/api/v1/files/550e8400-e29b-41d4-a716-446655440000
```

Expected response:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "original_name": "test.txt",
  "original_size": 47,
  "total_chunks": 1,
  "total_shards": 3,
  "n": 3,
  "k": 2,
  "status": "UPLOADED",
  "health_status": {
    "healthy_shards": 3,
    "corrupted_shards": 0,
    "missing_shards": 0,
    "total_shards": 3,
    "health_percent": 100,
    "recoverable": true
  }
}
```

### Download the file
```bash
curl -o downloaded.txt http://localhost:8080/api/v1/files/550e8400-e29b-41d4-a716-446655440000/download
```

### Verify downloaded file matches original
```bash
diff test.txt downloaded.txt
# Should output nothing if files match
```

## Step 6: Test Shard Map Endpoints

### Get shard map for file
```bash
curl http://localhost:8080/api/v1/shards/file/550e8400-e29b-41d4-a716-446655440000
```

Expected response shows all shards:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "original_name": "test.txt",
  "original_size": 47,
  "total_chunks": 1,
  "n": 3,
  "k": 2,
  "shard_size": 24,
  "status": "UPLOADED",
  "shards": [
    {
      "shard_id": "...",
      "chunk_index": 0,
      "shard_index": 0,
      "type": "DATA",
      "remote_id": "s3-mock-remote-id",
      "provider": "awsS3",
      "checksum_sha256": "...",
      "status": "HEALTHY"
    },
    ...
  ]
}
```

## Step 7: Test Erasure Coding Resilience

The system should be able to reconstruct files even with missing shards (as long as K shards are available).

This would require:
1. Simulating shard deletion or corruption
2. Attempting file download
3. Verifying successful reconstruction

**Note:** Currently the mock adapters return placeholder data, so full end-to-end testing requires implementing actual cloud storage operations.

## Step 8: Database Verification

Check that data is properly stored:

```bash
psql -U postgres -d cs464_project

# Check files table
SELECT id, original_name, original_size, total_chunks, status FROM files;

# Check shards table
SELECT id, file_id, chunk_index, shard_index, shard_type, provider, status FROM shards;

# Exit
\q
```

## Common Issues

### Database Connection Failed
- Verify PostgreSQL is running: `pg_isready`
- Check DATABASE_URL in `.env` file
- Verify database exists: `psql -U postgres -l`

### Server Won't Start
- Check if port 8080 is already in use: `lsof -i :8080`
- Verify migrations were applied
- Check logs for specific error messages

### Upload Fails
- Verify file size is reasonable
- Check k <= n constraint
- Ensure providers exist in registry
- Check server logs for detailed errors

### Download Returns Wrong Data
- Currently expected with mock adapters
- Mock adapters return placeholder data, not actual uploaded content
- Implement real cloud storage adapters for full functionality

## Next Steps for Production

1. **Implement Real Cloud Adapters**
   - AWS S3 SDK integration
   - Google Drive API integration
   - Add authentication and configuration

2. **Add Unit Tests**
   - Service layer tests
   - Handler tests
   - Repository tests

3. **Add Integration Tests**
   - End-to-end upload/download flows
   - Erasure coding verification
   - Shard recovery scenarios

4. **Performance Testing**
   - Large file uploads (GB scale)
   - Concurrent operations
   - Recovery time measurements

5. **Security Hardening**
   - Input validation
   - Rate limiting
   - Authentication/Authorization
   - HTTPS/TLS

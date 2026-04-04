# CS464 Proposal

## Description

Distributed High-availability File System, implementing Reed-Solomon sharding algorithm across multiple cloud storage providers. 

Videos on Reed-Solomon  
[What are Reed-Solomon Codes? How computers recover lost data](https://www.youtube.com/watch?v=1pQJkt7-R4Q)  
[Erasure Coding in Action: How Big Tech Saves Petabytes of Data](https://www.youtube.com/watch?v=VUxXH4uo4AY)

## Diagram

[https://excalidraw.com/\#json=qs5T3z3wE9JKbdhoW5hZd,ETaRTLYK\_HRm9ZLg2lyvJg](https://excalidraw.com/#json=qs5T3z3wE9JKbdhoW5hZd,ETaRTLYK_HRm9ZLg2lyvJg)  
![][image1]

## Flow

Uploading Files

1. The user sends a POST request containing the encrypted file to the **Shard Orchestrator**.  
2. The Orchestrator passes the file to the **Sharding/Reconstructing Service** (in this case sharding), which splits it into data and parity shards (using Reed-Solomon), and returns the generated pieces back to the Orchestrator.  
3. The Orchestrator sends the shard details to the **Shard Map** (database). The database records the unique IDs and assigned cloud locations for each piece, and sends a confirmation back to the Orchestrator.  
4. Finally, the Orchestrator forwards the shards to the **Adapter Service** (the Translator). The Adapter converts the upload commands and streams the shards to their assigned **Cloud Providers**.

Downloading Files

1. The client sends a GET request for a specific file to the **Shard Orchestrator**.  
2. The Orchestrator queries the **Shard Map** (database) to retrieve the unique data IDs and cloud locations for all shards belonging to that file, bringing that map back to the Orchestrator.  
3. The Orchestrator passes those identifiers to the **Adapter Service**. The Adapter translates the request, fetches the physical shards from the **Cloud Providers**, and returns them to the Orchestrator.  
4. The Orchestrator sends the downloaded pieces to the **Sharding/Reconstructing Service** (in this case reconstructing), which merges the shards back into the original file and hands it back to the Orchestrator.  
5. The Orchestrator serves the fully reconstructed file back to the client.

## Services

### Shard Orchestrator

1. Upload coordination  
  When a user uploads a file, the client sends a batch of encrypted shards to the orchestrator. It is responsible for deciding which shard goes to which provider — balancing across providers based on remaining quota — and then dispatching all uploads in parallel using Promise.all. It also writes the resulting shard map (shard ID → provider → path) into the shard metadata store so the system can find them again later.  
     
2. Download coordination  
  On retrieval, the orchestrator queries the shard metadata store for the shard map, then fires parallel fetch requests to each provider. Critically, because Reed-Solomon only requires K-of-N shards to reconstruct a file, the orchestrator implements an early exit strategy — it resolves as soon as the minimum required shards arrive, discarding the slower providers entirely. This minimises download latency without downloading more data than necessary.  
     
3. Health-aware decisions  
  During both upload and download, the orchestrator checks the health status of each provider (sourced from the shard metadata store) before routing to it. A provider flagged as degraded or at capacity is skipped, and the orchestrator selects an alternative. This is what gives the system its fault-tolerance at the routing level, before the Reed-Solomon layer even needs to intervene.  
     
4. Repair orchestration  
  When the Background Worker detects a missing or corrupt shard, it does not fetch or reconstruct data itself — it delegates to the orchestrator. The orchestrator downloads enough healthy shards to reconstruct the missing one via Reed-Solomon, then uploads the repaired shard to a different healthy provider, and finally updates the shard map in the shard metadata store to reflect the new location.

### Sharding/Reconstruction Service

The Sharding/Reconstruction Service provides two main operations

* Sharding (Encoding), it splits an uploaded file into K data shards and generate T (N \- K) parity shards using the Reed-Solomon algorithm  
* Reconstruction (Decoding), it recovers the original file from any available K shards (data or parity) using matrix inversion if required.

Both operations require the same mathematical foundation, which is an encoding matrix constructed over a Galois Field of 2^8. This service is stateless in operation, it merely encodes or decodes the input. This service uses synchronous HTTP/REST for its communication as the Orchestrator must receive complete results before proceeding to storage distribution

**API Endpoints**

1. POST /shard  
   Shards a file into N total shards (K data \+ T parity)  
     
   Request body  
* file\_id (uuid string)  
* file\_data (base64 encoded binary)  
* N (int)  
* K (int)

  Response body

* File ID (uuid string)  
* Shards array:  
  * shard\_index (int)  
  * shard\_type (“data” | “parity”)  
  * shard\_data (base64 encoded binary)  
* Metadata:  
  * N (int)  
  * K (int)  
  * original\_size (int)  
  * shard\_size (int)

2. POST /reconstruct  
   Reconstructs the original file from available shards  
     
   Request body  
* file\_id (uuid string)  
* N (int)  
* K (int)  
* Available Shards array:  
  * shard\_index (int)  
  * shard\_type (“data” | “parity”)  
  * shard\_data (base64 encoded binary)


  Response body

* file\_id (uuid string)  
* reconstructed\_file (base64 encoded binary)  
* Metadata  
  * original\_size (int)  
  * shards\_used (int)  
  * reconstruction\_method (“data” | “parity”)

### Shard Map Service

The Shard Map is used as the authoritative registry for all file-to-shard mappings and their locations. It stores the immutable file parameters (N, K, original size) and mutable shard location information. This service is implemented separately for modular design and domain isolation.

**API Endpoints**

* POST /files  
  Creates a new file entry with related metadata  
  **Request body**  
* File ID (uuid string)  
* N (int)  
* K (int)  
* original\_size (int)  
* shard\_size (int)  
* Shard distribution array:  
  * shard\_id (uuid string)  
  * shard\_index (int)  
  * provider (string)  
  * shard\_type (“data” | “parity”)  
  * remote\_id (string)

* GET /files/{file\_id}  
  Retrieves complete shard metadata for reconstruction planning  
    
* PATCH /files/{file\_id}/shards/{shards\_index}  
  Updates shard status or provider

### Adapter Service

Acts as a structural bridge between the application's core logic and the varying APIs of external cloud providers (AWS S3, Google Drive, and OneDrive).  
During the research phase, we explored unified API platforms like **Apideck** to simplify these integrations; however, we found that they do not provide the necessary support for **AWS S3**, therefore we decided to implement our own.

**API Endpoints**

1. POST /shards/upload  
   Uploads a shard with related metadata  
   Request body  
* shard\_id (uuid string)  
* provider (string)  
* size (int)  
* file\_data (binary stream)  
  Response body  
* remote\_id (string) \- Unique ID returned by the Cloud SDK  
* checksum\_sha256 \- Fingerprint of the shard after successful upload  
2. GET /shards/{remote\_id}  
   Retrieves a binary shard stream from a provider for file reconstruction  
   Query parameters  
* provider (string)

	Response

* binary\_stream (application/octet-stream)  
3. DELETE /shards/{remote\_id}  
   Removes a shard from a provider; triggered by the Saga Pattern during an atomic rollback or manual file deletion  
   Query parameters  
* remote\_id (string)

	Response

* status  
4. GET /providers/quota  
   Normalizes storage metadata across all connected accounts for the dashboard  
   Query parameters  
* provider (string)

Response

* total\_bytes (int)  
* used\_bytes (int)  
* status

## Cloud Providers

To ensure maximum reliability and security, the system communicates with each provider via their official Golang SDKs.

### Google Drive

* SDK: [google.golang.org/api/drive/v3](http://google.golang.org/api/drive/v3)  
* Communication Method  
  * REST-based Google Drive API  
* Authentication  
  * Managed via OAuth 2.0

### AWS S3

* SDK: github.com/aws/aws-sdk-go-v2  
* Communication Method  
  * Utilize the s3/transfermanager package  
* Authentication  
  * Managed via OAuth 2.0  
* Reference: [Amazon S3 examples using SDK for Go V2 \- AWS SDK Code Examples](https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html)

### OneDrive

* SDK: [goh-chunlin/go-onedrive: This is a Golang client library for accessing the Microsoft OneDrive REST API.](https://github.com/goh-chunlin/go-onedrive)  
  OR  
  https://pkg.go.dev/github.com/tonimelisma/onedrive-client/pkg/onedrive  
* Communication Method  
  * REST-based client that wraps the Microsoft Graph API  
* Authentication  
  * Managed via OAuth 2.0

## Frontend

### Framework & Core Libraries

* Next.js (`next@16.1.6`) — React-based framework used for routing, rendering, and overall application structure.  
* React (`react@19.2.3`, `react-dom@19.2.3`) — Core UI library for building components and interactive interfaces.

### Language & Tooling

* TypeScript (`typescript@^5`) — Provides static typing and improved developer experience.

### Styling & UI Components

* Tailwind CSS (`tailwindcss@^4`) — Utility-first CSS framework for styling.  
* PostCSS (`@tailwindcss/postcss`) — CSS processing pipeline used alongside Tailwind.  
* shadcn/ui — Component system/configuration (based on project structure such as `components/ui/*` and `components.json`).  
* Radix UI (`radix-ui`) — Accessible UI primitives used by the component layer.

### State Management & Data Fetching

* Zustand (`zustand@^5.0.11`) — Lightweight client-side state management.  
* TanStack React Query (`@tanstack/react-query@^5.90.21`) — Server-state management and caching for data fetching.

[image1]: <data:image/png;base64,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>
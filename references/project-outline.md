# PROJECT BRIEF: Distributed Cloud Storage (Frontend Development)

## OVERVIEW
Build the frontend for a distributed cloud storage system that combines multiple free cloud storage accounts (Google Drive, AWS S3, Dropbox) into one fault-tolerant storage solution. Files are encrypted client-side, split into shards using Reed-Solomon erasure coding, and distributed across providers so no single provider can access or lose user data.

Target users: Developers who want transparency, control, and privacy in their cloud storage.

---

## CORE CONCEPT
Think "RAID for cloud storage" - split files across multiple cloud providers with encryption and redundancy. Users can lose access to 1-2 providers and still recover their files.

---

## TECHNICAL STACK

**Framework:** Next.js 15 (App Router)
**Language:** TypeScript (strict mode)
**Styling:** Tailwind CSS + shadcn/ui components
**State Management:** Zustand (for upload progress) + React Query (for API calls)
**Authentication:** NextAuth.js (OAuth with Google)
**File Handling:** Web Crypto API for encryption, File API for uploads

---

## REED-SOLOMON CONFIGURATION (Simplified MVP)

**Configuration:** (6,4) 
- 6 total shards (4 data + 2 parity)
- Need any 4 shards to reconstruct file
- 50% storage overhead (2GB file → 3GB stored)
- Distribute across 2 providers initially (Google Drive + AWS S3)

**Note:** Display this visibly in UI so developers understand the redundancy mechanism.

---

## USER WORKFLOW (Step-by-Step)

### **1. Landing Page (Unauthenticated)**
- Hero section explaining the value proposition:
  - "Combine multiple cloud providers into one fault-tolerant drive"
  - "Zero-knowledge encryption - no provider can read your files"
  - "Survives any single provider outage"
- Primary CTA: "Get Started" button
- Secondary info: Feature highlights (encryption, redundancy, free tier aggregation)

### **2. Authentication**
User clicks "Get Started" → Redirect to auth page with two options:
- **Primary:** "Sign up with Google" (also connects Google Drive automatically)
- **Secondary:** "Sign up with Email + Password"

Most users choose Google OAuth for convenience.

**After successful auth:**
- Account created in database
- If OAuth: First provider (Google Drive) auto-connected
- Redirect to provider setup page

### **3. Provider Connection (Onboarding)**
**Goal:** Connect at least 2 cloud providers for redundancy

**UI Layout:**
- Title: "Connect Cloud Providers"
- Subtitle: "You need at least 2 providers for fault tolerance. Connect 2-3 for best redundancy."
- Provider cards displayed in a grid:

**Each provider card shows:**
- Provider icon + name (e.g., "Google Drive")
- Connection status: "Connected ✅" or "Not Connected"
- Free tier capacity: "15GB free"
- Current usage: "0 / 15 GB"
- Action button: "Connect" (if not connected) or "Disconnect" (if connected)

**Validation:**
- "Continue" button disabled until minimum 2 providers connected
- Progress indicator: "1/2 minimum providers" → "2/2 providers ✅"

**After connecting 2+ providers:**
- "Continue" button becomes active
- User proceeds to dashboard

### **4. Main Dashboard (Empty State)**
**First-time user sees:**
- Sidebar navigation (persistent):
  - 📁 My Files (active)
  - ☁️ Cloud Providers
  - ⚙️ Settings
- Main content area:
  - Empty state illustration
  - Message: "You haven't uploaded any files yet"
  - Storage summary: "Combined storage: 20GB across 2 providers"
  - Large "Upload Your First File" button

**Storage usage widget (in sidebar or top):**
- Visual bar showing total usage: "0 / 20 GB"
- Breakdown by provider:
  - Google Drive: 0 / 15 GB
  - AWS S3: 0 / 5 GB

**Provider health status (in sidebar):**
- ● Google Drive ✅ (green)
- ● AWS S3 ✅ (green)

### **5. File Upload Flow**

**Trigger:** User clicks "Upload File" button or drags file onto dashboard

**Upload Modal (Step 1):**
- File dropzone: "Drag & drop file here or click to browse"
- After file selected:
  - Show file name + size: "backup.sql (2.4 GB)"
  - Checkbox (checked by default): "☑ Enable zero-knowledge encryption"
  - Info text: "Your file will be encrypted in your browser before upload. No cloud provider can read the contents."
  - Distribution preview:
    - "This file will be split into 6 shards:"
    - "• 3 shards → Google Drive"
    - "• 3 shards → AWS S3"
    - "• You need any 4 shards to recover the file"
  - Action buttons: [Cancel] [Upload File →]

**Upload Progress (Step 2):**
After clicking "Upload File", show real-time progress:
```
Uploading backup.sql

Step 1: Encrypting file...          ✅ Complete
Step 2: Creating shards (6,4)...    ✅ Complete  
Step 3: Uploading shards...         ⏳ 67%

┌─────────────────────────────────┐
│ ● Shard 1 → Google Drive    ✅  │
│ ● Shard 2 → Google Drive    ✅  │
│ ● Shard 3 → Google Drive    ✅  │
│ ● Shard 4 → AWS S3          ⏳  │ [████████░░] 80%
│ ● Shard 5 → AWS S3          📤  │ [███░░░░░░░] 30%
│ ● Shard 6 → AWS S3          📤  │ [██░░░░░░░░] 20%
└─────────────────────────────────┘

Overall Progress: [████████░░] 67%

[Cancel Upload]
```

**Icons legend:**
- ✅ = Complete
- ⏳ = Uploading (current)
- 📤 = Queued
- ❌ = Failed

**After upload completes:**
- Show success message: "✅ backup.sql uploaded successfully"
- Auto-close modal after 2 seconds
- Redirect to file list (file now appears in dashboard)

### **6. File List (Dashboard with Files)**

**Layout:**
- Sidebar (same as before)
- Main content:
  - Search bar: "🔍 Search files..."
  - Sort/filter options: "Sort by: Date ▼" "Filter: All files ▼"
  - File cards in a grid (or list view)

**Each file card displays:**
```
┌────────────────────────────────────┐
│ 📄 backup.sql                      │
│ 2.4 GB • Uploaded 2 hours ago      │
│                                    │
│ Shards: 6/6 ✅ • Health: 100%      │
│                                    │
│ [Download] [Details ▼] [Delete]    │
└────────────────────────────────────┘
```

**Health indicator colors:**
- 100% (6/6 or 5/6 shards): Green ✅
- 67-99% (4/6 shards): Yellow ⚠️
- <67% (<4/6 shards): Red ❌ "Cannot recover"

**File card with issues:**
```
┌────────────────────────────────────┐
│ 📄 old_backup.sql                  │
│ 1.8 GB • Uploaded 3 days ago       │
│                                    │
│ Shards: 5/6 ⚠️ • Health: 83%       │
│ Note: 1 shard unreachable          │
│                                    │
│ [Download] [Details ▼] [Repair]    │
└────────────────────────────────────┘
```

### **7. File Download Flow**

**Trigger:** User clicks "Download" button on a file card

**Download Modal:**
```
Downloading backup.sql

Step 1: Fetching shards...          ⏳ 50%
Step 2: Reconstructing file...      ⏸️ Waiting
Step 3: Decrypting...               ⏸️ Waiting

┌─────────────────────────────────┐
│ ✅ Shard 1 from Google Drive    │
│ ✅ Shard 2 from Google Drive    │
│ ⏳ Shard 3 from AWS S3 (45%)    │
│ 📥 Shard 4 from AWS S3          │
│ ⏸️ Shard 5 (not needed)         │
│ ⏸️ Shard 6 (not needed)         │
└─────────────────────────────────┘

Note: Only 4/6 shards needed for recovery

Overall Progress: [█████░░░░░] 50%
```

**Key UX insight:** 
- Download shards in parallel
- Only wait for first 4 shards (not all 6)
- Show that extra shards aren't needed (⏸️ paused)

**After download completes:**
- File automatically saves to browser's Downloads folder
- Success message: "✅ backup.sql downloaded successfully"
- Show file checksum: "SHA-256: a3f2...8b1c (verified)"

### **8. File Details View**

**Trigger:** User clicks "Details ▼" on a file card

**Layout:** Full-page view with back button
```
← Back to Files

backup.sql
2.4 GB • Uploaded Feb 19, 2024 2:30 PM

┌─────────────────────────────────────┐
│ 🔐 Encryption                       │
│ Algorithm: AES-256-GCM              │
│ Encryption Key: ••••••••••••••••    │
│ [Show Key] [Copy Key]               │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ 🧩 Shard Distribution (6,4)         │
│                                     │
│ Data Shards (4):                    │
│ • Shard 1: Google Drive ✅          │
│ • Shard 2: Google Drive ✅          │
│ • Shard 3: AWS S3 ✅                │
│ • Shard 4: AWS S3 ✅                │
│                                     │
│ Parity Shards (2):                  │
│ • Shard 5: Google Drive ✅          │
│ • Shard 6: AWS S3 ✅                │
│                                     │
│ Health: 6/6 shards (100%)           │
│ Status: ✅ Can be fully recovered   │
│ [Test Download]                     │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ 📊 Metadata                         │
│ Original SHA-256: a3f2...8b1c       │
│ Total Shards: 6                     │
│ Storage Used: 3.6 GB (1.5x file)    │
│ Created: Feb 19, 2024 2:30 PM       │
│ Last Verified: 2 hours ago          │
└─────────────────────────────────────┘

[Download File] [Delete File]
```

**Developer-friendly details:**
- Show exact shard locations
- Display checksums for verification
- Expose storage overhead calculation
- Show last verification timestamp

### **9. Cloud Providers Management**

**Route:** `/providers`

**Layout:**
```
Cloud Providers

┌───────────────────────────────────┐
│ ☁️ Google Drive                   │
│                                   │
│ Status: ✅ Connected              │
│ Account: user@gmail.com           │
│ Quota: 8.2 / 15 GB (55%)          │
│ [████████░░░░░░░░] 55%            │
│                                   │
│ Shards Stored: 142 shards         │
│ Last Health Check: 5 mins ago ✅  │
│                                   │
│ [Disconnect] [Refresh Quota]      │
└───────────────────────────────────┘

┌───────────────────────────────────┐
│ ☁️ AWS S3                         │
│                                   │
│ Status: ✅ Connected              │
│ Account: my-bucket-name           │
│ Quota: 3.1 / 5 GB (62%)           │
│ [████████████░░░░] 62%            │
│                                   │
│ Shards Stored: 138 shards         │
│ Last Health Check: 5 mins ago ✅  │
│                                   │
│ [Disconnect] [Refresh Quota]      │
└───────────────────────────────────┘

[+ Add New Provider]
```

**Provider card with error:**
```
┌───────────────────────────────────┐
│ ☁️ Dropbox                        │
│                                   │
│ Status: ⚠️ Connection Issue       │
│ Last Error: "Rate limit exceeded" │
│ Occurred: 2 hours ago             │
│                                   │
│ [Reconnect] [View Logs]           │
└───────────────────────────────────┘
```

### **10. Settings Page**

**Route:** `/settings`

**Sections:**
- **Account Settings:**
  - Email: user@gmail.com
  - Auth method: Google OAuth
  - [Change Password] (if email auth)
  - [Delete Account]

- **Redundancy Configuration:**
  - Current: (6,4) - Balanced
  - [Change Configuration]
  - Warning: "Changing this will only affect new uploads"

- **Storage Preferences:**
  - Default encryption: On/Off toggle
  - Auto-delete old files: Off (with date picker if enabled)

- **Advanced:**
  - Export all file metadata (JSON)
  - View encryption keys
  - API access (future feature)

---

## UI COMPONENTS TO BUILD

### **Core Components:**

1. **FileUploadModal.tsx**
   - File dropzone
   - Encryption toggle
   - Upload progress with per-shard tracking
   - Server-Sent Events integration for real-time progress

2. **FileCard.tsx**
   - File name, size, timestamp
   - Shard health indicator (6/6 ✅)
   - Action buttons (Download, Details, Delete)
   - Visual health badge (green/yellow/red)

3. **ProviderCard.tsx**
   - Provider icon + name
   - Connection status
   - Quota usage bar
   - Connect/Disconnect buttons
   - Health check status

4. **ShardProgressBar.tsx**
   - Individual shard upload/download progress
   - Status icons (✅⏳📤❌⏸️)
   - Provider label
   - Percentage indicator

5. **HealthBadge.tsx**
   - Color-coded badge (green/yellow/red)
   - Shard count display (6/6, 5/6, etc.)
   - Tooltip with details

6. **Sidebar.tsx**
   - Navigation links
   - Storage usage summary
   - Provider health indicators
   - User profile dropdown

### **Utility Components:**

7. **EmptyState.tsx**
   - Illustration
   - Message
   - CTA button

8. **LoadingSpinner.tsx**
   - Animated spinner for async operations

9. **Toast.tsx**
   - Success/error notifications
   - Auto-dismiss

---

## STATE MANAGEMENT

### **Zustand Stores:**

**uploadStore.ts:**
```typescript
interface UploadState {
  activeUploads: Map<string, UploadProgress>;
  addUpload: (fileId: string) => void;
  updateShardProgress: (fileId: string, shardIndex: number, progress: number) => void;
  completeUpload: (fileId: string) => void;
  cancelUpload: (fileId: string) => void;
}

interface UploadProgress {
  fileId: string;
  filename: string;
  totalSize: number;
  stage: 'encrypting' | 'sharding' | 'uploading' | 'complete';
  shards: ShardProgress[];
  overallProgress: number;
}

interface ShardProgress {
  index: number;
  provider: string;
  status: 'pending' | 'uploading' | 'complete' | 'failed';
  progress: number; // 0-100
  uploadedBytes: number;
}
```

**providerStore.ts:**
```typescript
interface ProviderStore {
  providers: Provider[];
  healthStatus: Map<string, ProviderHealth>;
  fetchProviders: () => Promise<void>;
  connectProvider: (providerName: string) => Promise<void>;
  disconnectProvider: (providerId: string) => Promise<void>;
  checkHealth: (providerId: string) => Promise<void>;
}

interface Provider {
  id: string;
  name: string;
  status: 'connected' | 'disconnected' | 'error';
  quotaUsed: number;
  quotaTotal: number;
  shardsStored: number;
  lastHealthCheck: Date;
}
```

### **React Query Keys:**
```typescript
// Fetch all files
['files']

// Fetch single file details
['files', fileId]

// Fetch connected providers
['providers']

// Fetch provider quota
['providers', providerId, 'quota']
```

---

## API ROUTES (Frontend consumes these)

**Authentication:**
- `GET /api/auth/google` - Initiate Google OAuth
- `GET /api/auth/google/callback` - OAuth callback
- `POST /api/auth/signup` - Email signup
- `POST /api/auth/login` - Email login

**Files:**
- `POST /api/upload` - Upload file (multipart/form-data)
- `GET /api/upload/[uploadId]/progress` - Server-Sent Events for upload progress
- `GET /api/files` - List all user files
- `GET /api/files/[id]` - Get file details
- `GET /api/files/[id]/download` - Download file (returns stream)
- `DELETE /api/files/[id]` - Delete file

**Providers:**
- `GET /api/providers` - List connected providers
- `POST /api/providers/connect` - Initiate provider OAuth
- `DELETE /api/providers/[id]` - Disconnect provider
- `GET /api/providers/[id]/health` - Check provider health
- `POST /api/providers/[id]/refresh-quota` - Refresh quota info

---

## ENCRYPTION (Client-Side)

**Encrypt file before upload:**
```typescript
async function encryptFile(file: File): Promise<{ encrypted: ArrayBuffer, key: CryptoKey }> {
  // Generate encryption key
  const key = await crypto.subtle.generateKey(
    { name: 'AES-GCM', length: 256 },
    true,
    ['encrypt', 'decrypt']
  );
  
  // Generate random IV (initialization vector)
  const iv = crypto.getRandomValues(new Uint8Array(12));
  
  // Read file
  const fileBuffer = await file.arrayBuffer();
  
  // Encrypt
  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    fileBuffer
  );
  
  return { encrypted, key, iv };
}
```

**Decrypt file after download:**
```typescript
async function decryptFile(
  encryptedBuffer: ArrayBuffer,
  key: CryptoKey,
  iv: Uint8Array
): Promise<ArrayBuffer> {
  return await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv },
    key,
    encryptedBuffer
  );
}
```

---

## DESIGN GUIDELINES

**Color Scheme (Developer-Friendly):**
- Primary: Blue (#3B82F6) - used for CTAs, links
- Success: Green (#10B981) - healthy status
- Warning: Yellow (#F59E0B) - degraded status
- Error: Red (#EF4444) - failed status
- Background: Dark mode preferred (#1F2937 dark, #F9FAFB light)

**Typography:**
- Headings: Inter or Geist Sans (bold)
- Body: Inter or Geist Sans (regular)
- Monospace: Geist Mono (for file paths, checksums)

**Component Library:**
Use shadcn/ui for:
- Button
- Card
- Dialog (for modals)
- Badge
- Progress
- Tooltip
- Dropdown Menu
- Tabs

**Icons:**
Use lucide-react:
- Upload: `<Upload />`
- Download: `<Download />`
- Cloud: `<Cloud />`
- Check: `<Check />`
- Alert: `<AlertTriangle />`
- Settings: `<Settings />`
- File: `<File />`

---

## RESPONSIVE DESIGN

**Desktop (>1024px):**
- Sidebar always visible (250px width)
- File cards in 3-column grid
- Upload modal: centered, 600px width

**Tablet (768px - 1024px):**
- Sidebar collapsible
- File cards in 2-column grid
- Upload modal: centered, 80% width

**Mobile (<768px):**
- Sidebar becomes bottom nav bar
- File cards in single column (list view)
- Upload modal: full-screen
- Shard progress: collapsed by default (tap to expand)

---

## KEY INTERACTIONS

**File Upload:**
1. Drag file onto dashboard OR click "Upload" button
2. Modal opens with file preview
3. User confirms upload
4. Progress shown in real-time (per-shard)
5. Success toast notification
6. Modal auto-closes, file appears in list

**File Download:**
1. Click "Download" on file card
2. Progress modal appears
3. Show shard fetching progress
4. Once 4/6 shards fetched, begin reconstruction
5. File saves to Downloads folder
6. Success toast notification

**Provider Connection:**
1. Click "Connect [Provider]" button
2. Redirect to provider's OAuth page
3. User approves permissions
4. Redirect back to app
5. Provider card updates to "Connected ✅"
6. Quota usage displayed

**Error Handling:**
1. If shard upload fails: Retry 3 times with exponential backoff
2. If provider is unreachable: Show error state on provider card
3. If file can't be recovered (<4 shards): Show red badge, disable Download
4. All errors show toast notifications with actionable messages

---

## ACCESSIBILITY

- All interactive elements keyboard-navigable
- ARIA labels on all buttons
- Focus indicators visible
- Color not the only indicator (use icons too)
- Screen reader announcements for upload progress
- Alt text on all images/icons

---

## PERFORMANCE OPTIMIZATION

- Lazy load file cards (virtualized list for >100 files)
- Debounce search input (300ms)
- Use React Query for API caching
- Compress large images in file previews
- Use Web Workers for encryption (prevent UI blocking)
- Server-Sent Events for real-time progress (not WebSocket overhead)

---

## TECHNICAL CONSTRAINTS

**File Size Limits:**
- Max file size: 5GB (Web Crypto API limitation)
- For larger files: Show warning, suggest splitting
- Browser memory limit: ~2GB (varies by browser)

**Browser Support:**
- Chrome 90+ (Web Crypto API)
- Firefox 88+
- Safari 14+
- Edge 90+
- No IE11 support (uses modern APIs)

**Network Requirements:**
- Minimum 1 Mbps upload speed
- Parallel uploads (6 simultaneous connections)
- Resume support (future feature)

---

## FOLDER STRUCTURE
```
app/
├── (auth)/
│   ├── login/page.tsx
│   ├── signup/page.tsx
│   └── onboarding/page.tsx
├── (dashboard)/
│   ├── layout.tsx (includes Sidebar)
│   ├── page.tsx (file list)
│   ├── files/[id]/page.tsx (file details)
│   ├── providers/page.tsx
│   └── settings/page.tsx
├── api/
│   ├── auth/[...nextauth]/route.ts
│   ├── upload/route.ts
│   ├── files/route.ts
│   └── providers/route.ts
├── components/
│   ├── ui/ (shadcn components)
│   ├── FileCard.tsx
│   ├── FileUploadModal.tsx
│   ├── ProviderCard.tsx
│   ├── ShardProgressBar.tsx
│   ├── Sidebar.tsx
│   └── HealthBadge.tsx
├── lib/
│   ├── encryption.ts (client-side crypto)
│   ├── api.ts (API client)
│   └── utils.ts
├── stores/
│   ├── uploadStore.ts
│   └── providerStore.ts
└── types/
    ├── file.ts
    ├── provider.ts
    └── upload.ts
```

---

## DELIVERABLES

**Phase 1 (MVP - 8 weeks):**
1. Authentication (Google OAuth)
2. Provider connection (Google Drive + AWS S3)
3. File upload with encryption + sharding
4. File list with health indicators
5. File download with reconstruction
6. Basic error handling

**Phase 2 (Polish - 4 weeks):**
7. File details view
8. Provider management page
9. Settings page
10. Responsive design
11. Accessibility improvements
12. Performance optimization

---

## SUCCESS CRITERIA

**Functional Requirements:**
- ✅ User can sign up and connect 2 providers
- ✅ User can upload a file and see it split into shards
- ✅ User can see shard health indicators
- ✅ User can download a file and verify it matches original
- ✅ System works with 1 provider temporarily offline

**Non-Functional Requirements:**
- ✅ Upload progress updates in real-time (<1s latency)
- ✅ UI responsive on desktop, tablet, mobile
- ✅ No console errors in production
- ✅ Accessible (WCAG 2.1 AA)
- ✅ Loading states for all async operations

---

## TONE & STYLE

**Developer-Focused:**
- Show technical details (don't hide complexity)
- Use precise terminology (shards, parity, checksum)
- Expose system internals (shard distribution, health checks)
- Minimal hand-holding (assume technical literacy)

**Visual Style:**
- Clean, modern, professional
- Dark mode by default (light mode toggle)
- Monospace fonts for technical info
- Status indicators everywhere (✅⚠️❌)
- Real-time feedback (progress bars, live updates)

---

## EXAMPLE MOCK DATA (For Testing Without Backend)
```typescript
const mockFiles = [
  {
    id: '1',
    filename: 'backup_prod_2024.sql',
    size: 2400000000, // 2.4 GB
    uploadedAt: new Date('2024-02-19T14:30:00'),
    shards: [
      { index: 1, provider: 'google_drive', status: 'healthy' },
      { index: 2, provider: 'google_drive', status: 'healthy' },
      { index: 3, provider: 'aws_s3', status: 'healthy' },
      { index: 4, provider: 'aws_s3', status: 'healthy' },
      { index: 5, provider: 'google_drive', status: 'healthy' },
      { index: 6, provider: 'aws_s3', status: 'healthy' },
    ],
    health: 100,
    encryptionKey: 'a3f29b8c...'
  },
  {
    id: '2',
    filename: 'old_backup.sql',
    size: 1800000000, // 1.8 GB
    uploadedAt: new Date('2024-02-16T10:00:00'),
    shards: [
      { index: 1, provider: 'google_drive', status: 'healthy' },
      { index: 2, provider: 'google_drive', status: 'unreachable' },
      { index: 3, provider: 'aws_s3', status: 'healthy' },
      { index: 4, provider: 'aws_s3', status: 'healthy' },
      { index: 5, provider: 'google_drive', status: 'healthy' },
      { index: 6, provider: 'aws_s3', status: 'healthy' },
    ],
    health: 83, // 5/6 shards
    encryptionKey: 'b4e3ac7d...'
  }
];

const mockProviders = [
  {
    id: '1',
    name: 'google_drive',
    displayName: 'Google Drive',
    status: 'connected',
    quotaUsed: 8200000000, // 8.2 GB
    quotaTotal: 15000000000, // 15 GB
    shardsStored: 142,
    lastHealthCheck: new Date('2024-02-19T16:25:00')
  },
  {
    id: '2',
    name: 'aws_s3',
    displayName: 'AWS S3',
    status: 'connected',
    quotaUsed: 3100000000, // 3.1 GB
    quotaTotal: 5000000000, // 5 GB
    shardsStored: 138,
    lastHealthCheck: new Date('2024-02-19T16:25:00')
  }
];
```

---

## START HERE

**Immediate Next Steps:**
1. Set up Next.js 15 project with TypeScript
2. Install dependencies: shadcn/ui, Zustand, React Query, lucide-react
3. Create folder structure
4. Build Sidebar component (navigation foundation)
5. Build empty state dashboard
6. Build FileUploadModal (without backend integration - mock progress)
7. Build FileCard component with mock data
8. Build ProviderCard component with mock data

**Then:**
9. Integrate NextAuth.js for authentication
10. Connect to backend API routes (when ready)
11. Implement real encryption using Web Crypto API
12. Add Server-Sent Events for upload progress
13. Polish UI and add animations
14. Test responsive design
15. Add accessibility features

---

BEGIN IMPLEMENTATION. Focus on developer UX - show technical details, real-time progress, and system transparency. Prioritize functionality over aesthetics initially, then polish.
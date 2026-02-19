export interface FileData {
  id: string;
  name: string;
  size: string;
  uploadedAt: string;
  shardsAvailable: number;
  shardsTotal: number;
}

export const mockFiles: FileData[] = [
  {
    id: '1',
    name: 'project_backup_v2.zip',
    size: '2.4 GB',
    uploadedAt: '2 hours ago',
    shardsAvailable: 6,
    shardsTotal: 6,
  },
  {
    id: '2',
    name: 'client_assets_2024.rar',
    size: '1.8 GB',
    uploadedAt: 'Yesterday',
    shardsAvailable: 5,
    shardsTotal: 6,
  },
  {
    id: '3',
    name: 'sensitive_docs_encrypted.pdf',
    size: '15 MB',
    uploadedAt: '3 days ago',
    shardsAvailable: 4,
    shardsTotal: 6,
  },
  {
    id: '4',
    name: 'family_photos_2023.zip',
    size: '4.2 GB',
    uploadedAt: '1 week ago',
    shardsAvailable: 3, // Critical/Unrecoverable
    shardsTotal: 6,
  },
];

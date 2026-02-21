export interface ProviderData {
  id: string;
  name: string;
  status: "connected" | "disconnected" | "error";
  used: string;
  total: string;
  percentage: number;
  shardCount: number;
  lastCheck: string;
}

export const mockProviders: ProviderData[] = [
  {
    id: 'google-drive',
    name: 'Google Drive',
    status: 'connected',
    used: '8.2 GB',
    total: '15 GB',
    percentage: 55,
    shardCount: 142,
    lastCheck: '5 mins ago',
  },
  {
    id: 'aws-s3',
    name: 'AWS S3',
    status: 'connected',
    used: '3.1 GB',
    total: '5 GB',
    percentage: 62,
    shardCount: 138,
    lastCheck: '5 mins ago',
  },
  {
    id: 'dropbox',
    name: 'Dropbox',
    status: 'error',
    used: '0 GB',
    total: '2 GB',
    percentage: 0,
    shardCount: 0,
    lastCheck: 'Failed 2h ago',
  },
];

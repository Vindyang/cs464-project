export interface ProviderData {
  providerId: string;
  displayName: string;
  status: "connected" | "degraded" | "disconnected" | "error";
  region: string;
  latencyMs: number;
  quotaUsedBytes: number;
  quotaTotalBytes: number;
  shardCount: number;
  lastHealthCheckAt: string;
}

export const mockProviders: ProviderData[] = [
  {
    providerId: 'googleDrive',
    displayName: 'Google Drive',
    status: 'connected',
    region: 'global',
    latencyMs: 45,
    quotaUsedBytes: 8804682956, // 8.2 GB
    quotaTotalBytes: 16106127360, // 15 GB
    shardCount: 142,
    lastHealthCheckAt: '5 mins ago',
  },
  {
    providerId: 'awsS3',
    displayName: 'AWS S3',
    status: 'connected',
    region: 'us-east-1',
    latencyMs: 24,
    quotaUsedBytes: 3328599654, // 3.1 GB
    quotaTotalBytes: 5368709120, // 5 GB
    shardCount: 138,
    lastHealthCheckAt: '5 mins ago',
  },
  {
    providerId: 'dropbox',
    displayName: 'Dropbox',
    status: 'error',
    region: 'eu-west',
    latencyMs: 112,
    quotaUsedBytes: 0,
    quotaTotalBytes: 2147483648, // 2 GB
    shardCount: 0,
    lastHealthCheckAt: 'Failed 2h ago',
  },
];

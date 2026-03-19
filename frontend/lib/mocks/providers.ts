export interface ProviderData {
  providerId: string;
  displayName: string;
  status: string;
  region: string;
  latencyMs: number;
  quotaUsedBytes: number;
  quotaTotalBytes: number;
  capabilities?: Record<string, unknown>;
  lastHealthCheckAt: string;
}

export const mockProviders: ProviderData[] = [
  {
    providerId: 'googleDrive',
    displayName: 'Google Drive',
    status: 'connected',
    region: 'global',
    latencyMs: 45,
    quotaUsedBytes: 8804682956,
    quotaTotalBytes: 16106127360,
    lastHealthCheckAt: '5 mins ago',
  },
  {
    providerId: 'awsS3',
    displayName: 'AWS S3',
    status: 'connected',
    region: 'us-east-1',
    latencyMs: 24,
    quotaUsedBytes: 3328599654,
    quotaTotalBytes: 5368709120,
    lastHealthCheckAt: '5 mins ago',
  },
];

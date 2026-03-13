import { create } from 'zustand';

export interface ShardProgress {
  index: number;
  providerId: string;
  status: 'pending' | 'uploading' | 'complete' | 'failed' | 'paused';
  progress: number; // 0-100
  uploadedBytes: number;
}

export interface UploadProgress {
  fileId: string;
  filename: string;
  totalSize: number;
  stage: 'encrypting' | 'sharding' | 'uploading' | 'complete';
  shards: ShardProgress[];
  overallProgress: number;
}

interface UploadState {
  activeUploads: Record<string, UploadProgress>;
  
  // Actions
  addUpload: (file: File) => string;
  updateShardProgress: (fileId: string, shardIndex: number, progress: number) => void;
  setUploadStage: (fileId: string, stage: UploadProgress['stage']) => void;
  completeUpload: (fileId: string) => void;
  cancelUpload: (fileId: string) => void;
}

export const useUploadStore = create<UploadState>((set, get) => ({
  activeUploads: {},

  addUpload: (file: File) => {
    const fileId = Math.random().toString(36).substring(7);
    const initialShards: ShardProgress[] = Array(6).fill(null).map((_, i) => ({
        index: i,
        providerId: i % 2 === 0 ? 'googleDrive' : 'awsS3', // Mock distribution
        status: 'pending',
        progress: 0,
        uploadedBytes: 0
    }));

    set((state) => ({
      activeUploads: {
        ...state.activeUploads,
        [fileId]: {
          fileId,
          filename: file.name,
          totalSize: file.size,
          stage: 'encrypting',
          shards: initialShards,
          overallProgress: 0,
        },
      },
    }));

    return fileId;
  },

  updateShardProgress: (fileId, shardIndex, progress) => {
    set((state) => {
        const upload = state.activeUploads[fileId];
        if (!upload) return state;

        const newShards = [...upload.shards];
        newShards[shardIndex] = {
            ...newShards[shardIndex],
            progress,
            status: progress === 100 ? 'complete' : 'uploading'
        };

        // Calculate overall progress
        const totalProgress = newShards.reduce((acc, shard) => acc + shard.progress, 0) / newShards.length;

        return {
            activeUploads: {
                ...state.activeUploads,
                [fileId]: {
                    ...upload,
                    shards: newShards,
                    overallProgress: totalProgress,
                    stage: totalProgress === 100 ? 'complete' : 'uploading'
                }
            }
        };
    });
  },

  setUploadStage: (fileId, stage) => {
      set((state) => {
          const upload = state.activeUploads[fileId];
          if (!upload) return state;
          return {
              activeUploads: {
                  ...state.activeUploads,
                  [fileId]: { ...upload, stage }
              }
          }
      })
  },

  completeUpload: (fileId) => {
    set((state) => {
        const { [fileId]: _, ...remaining } = state.activeUploads;
        return { activeUploads: remaining };
    });
  },

  cancelUpload: (fileId) => {
    set((state) => {
        const { [fileId]: _, ...remaining } = state.activeUploads;
        return { activeUploads: remaining };
    });
  },
}));

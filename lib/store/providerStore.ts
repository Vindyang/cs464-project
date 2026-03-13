import { create } from 'zustand';
import { mockProviders, ProviderData } from '@/lib/mocks/providers';

interface ProviderStore {
  providers: ProviderData[];
  
  // Actions
  fetchProviders: () => void;
  connectProvider: (providerId: string) => void;
  disconnectProvider: (providerId: string) => void;
}

export const useProviderStore = create<ProviderStore>((set) => ({
  providers: mockProviders,

  fetchProviders: () => {
    // In a real app, this would be an async fetch
    set({ providers: mockProviders });
  },

  connectProvider: (providerId) => {
    set((state) => ({
      providers: state.providers.map((p) =>
        p.providerId === providerId ? { ...p, status: 'connected' } : p
      ),
    }));
  },

  disconnectProvider: (providerId) => {
    set((state) => ({
      providers: state.providers.map((p) =>
        p.providerId === providerId ? { ...p, status: 'disconnected', quotaUsedBytes: 0, shardCount: 0 } : p
      ),
    }));
  },
}));

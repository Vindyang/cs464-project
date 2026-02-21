import { create } from 'zustand';
import { mockProviders, ProviderData } from '@/lib/mocks/providers';

interface ProviderStore {
  providers: ProviderData[];
  
  // Actions
  fetchProviders: () => void;
  connectProvider: (id: string) => void;
  disconnectProvider: (id: string) => void;
}

export const useProviderStore = create<ProviderStore>((set) => ({
  providers: mockProviders,

  fetchProviders: () => {
    // In a real app, this would be an async fetch
    set({ providers: mockProviders });
  },

  connectProvider: (id) => {
    set((state) => ({
      providers: state.providers.map((p) =>
        p.id === id ? { ...p, status: 'connected' } : p
      ),
    }));
  },

  disconnectProvider: (id) => {
    set((state) => ({
      providers: state.providers.map((p) =>
        p.id === id ? { ...p, status: 'disconnected', percentage: 0, used: '0 GB', shardCount: 0 } : p
      ),
    }));
  },
}));

import { create } from 'zustand';
import { ProviderData } from '@/lib/mocks/providers';
import { getProviders } from '@/lib/api/providers';

interface ProviderStore {
  providers: ProviderData[];
  isLoading: boolean;
  error: string | null;

  fetchProviders: () => Promise<void>;
  connectProvider: (providerId: string) => void;
  disconnectProvider: (providerId: string) => void;
}

export const useProviderStore = create<ProviderStore>((set) => ({
  providers: [],
  isLoading: false,
  error: null,

  fetchProviders: async () => {
    set({ isLoading: true, error: null });
    try {
      const data = await getProviders();
      set({ providers: data, isLoading: false });
    } catch {
      set({ error: 'Failed to load providers', isLoading: false });
    }
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
        p.providerId === providerId ? { ...p, status: 'disconnected', quotaUsedBytes: 0 } : p
      ),
    }));
  },
}));

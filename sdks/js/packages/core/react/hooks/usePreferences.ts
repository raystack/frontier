import { useState, useCallback, useEffect } from 'react';
import { V1Beta1Preference } from '~/api-client';
import { useFrontier } from '../contexts/FrontierContext';

export interface UsePreferences {
  preferences: V1Beta1Preference[];
  isLoading: boolean;
  isFetching: boolean;
  status: 'idle' | 'fetching' | 'loading';
  fetchPreferences: () => Promise<V1Beta1Preference[] | undefined>;
  updatePreferences: (
    preferences: V1Beta1Preference[]
  ) => Promise<V1Beta1Preference[] | undefined>;
}

export function usePreferences(): UsePreferences {
  const { client } = useFrontier();
  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const [status, setStatus] = useState<UsePreferences['status']>('idle');

  const fetchPreferences = useCallback(async () => {
    try {
      setStatus('fetching');
      const response =
        await client?.frontierServiceListCurrentUserPreferences();
      const data = response?.data.preferences || [];
      setPreferences(data);
      return data;
    } catch (err) {
      console.error(
        'frontier:sdk:: There is problem with fetching user preferences'
      );
      console.error(err);
    } finally {
      setStatus('idle');
    }
    return [];
  }, [client]);

  const updatePreferences = useCallback(
    async (preferences: V1Beta1Preference[]) => {
      try {
        setStatus('loading');
        const response =
          await client?.frontierServiceCreateCurrentUserPreferences({
            bodies: preferences
          });
        const data = response?.data?.preferences ?? [];
        setPreferences(data);
        return data;
      } catch (err) {
        console.error(
          'frontier:sdk:: There is problem with updating user preferences'
        );
        console.error(err);
      } finally {
        setStatus('idle');
      }
      return [];
    },
    [client]
  );

  useEffect(() => {
    fetchPreferences();
  }, [fetchPreferences]);

  return {
    preferences,
    status,
    isLoading: status === 'loading',
    isFetching: status === 'fetching',
    fetchPreferences,
    updatePreferences
  };
}

import { useState, useCallback, useEffect } from 'react';
import { V1Beta1Preference } from '../../api-client';
import { useFrontier } from '../contexts/FrontierContext';

type Preferences = Record<string, V1Beta1Preference>;

export interface UsePreferences {
  preferences: Preferences;
  isLoading: boolean;
  isFetching: boolean;
  status: 'idle' | 'fetching' | 'loading';
  fetchPreferences: () => Promise<V1Beta1Preference[] | undefined>;
  updatePreferences: (
    preferences: V1Beta1Preference[]
  ) => Promise<V1Beta1Preference[] | undefined>;
}

function getFormattedData(preferences: V1Beta1Preference[] = []): Preferences {
  return preferences.reduce((acc: Preferences, preference) => {
    if (preference?.name) acc[preference.name] = preference;
    return acc;
  }, {});
}

export function usePreferences(): UsePreferences {
  const { client, user } = useFrontier();
  const [preferences, setPreferences] = useState<Preferences>({});
  const [status, setStatus] = useState<UsePreferences['status']>('fetching');

  const fetchPreferences = useCallback(async () => {
    try {
      setStatus('fetching');
      const response =
        await client?.frontierServiceListCurrentUserPreferences();
      const data = response?.data.preferences || [];
      setPreferences(getFormattedData(data));
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
        setPreferences(getFormattedData(data));
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
    if (user?.id) {
      fetchPreferences();
    }
  }, [fetchPreferences, user?.id]);

  return {
    preferences,
    status,
    isLoading: status === 'loading',
    isFetching: status === 'fetching',
    fetchPreferences,
    updatePreferences
  };
}

import { useMemo } from 'react';
import { useQuery, useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { create } from '@bufbuild/protobuf';
import { FrontierServiceQueries, CreateCurrentUserPreferencesRequestSchema } from '@raystack/proton/frontier';

type Preference = {
  name?: string;
  value?: string;
  [key: string]: any;
};

type Preferences = Record<string, Preference>;

export interface UsePreferences {
  preferences: Preferences;
  isLoading: boolean;
  isFetching: boolean;
  status: 'idle' | 'fetching' | 'loading';
  fetchPreferences: () => void;
  updatePreferences: (
    preferences: Preference[]
  ) => Promise<void>;
}

function getFormattedData(preferences: Preference[] = []): Preferences {
  return preferences.reduce((acc: Preferences, preference) => {
    if (preference?.name) acc[preference.name] = preference;
    return acc;
  }, {});
}

export function usePreferences({
  autoFetch = true
}: {
  autoFetch?: boolean;
} = {}): UsePreferences {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const {
    data,
    isLoading: isFetchingPreferences,
    refetch
  } = useQuery(
    FrontierServiceQueries.listCurrentUserPreferences,
    {},
    { enabled: autoFetch }
  );

  const preferences = useMemo(
    () => getFormattedData(data?.preferences ?? []),
    [data?.preferences]
  );

  const {
    mutateAsync: updatePreferencesMutation,
    isPending: isUpdatingPreferences
  } = useMutation(FrontierServiceQueries.createCurrentUserPreferences, {
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.listCurrentUserPreferences,
          transport,
          input: {},
          cardinality: 'finite'
        })
      });
    },
    onError: (err) => {
      console.error(
        'frontier:sdk:: There is problem with updating user preferences'
      );
      console.error(err);
    }
  });

  const updatePreferences = async (preferences: Preference[]) => {
    try {
      const req = create(CreateCurrentUserPreferencesRequestSchema, {
        bodies: preferences
      });
      await updatePreferencesMutation(req);
    } catch (err) {
      console.error(
        'frontier:sdk:: There is problem with updating user preferences'
      );
      console.error(err);
      throw err;
    }
  };

  const fetchPreferences = () => {
    refetch();
  };

  const status: UsePreferences['status'] = isUpdatingPreferences
    ? 'loading'
    : isFetchingPreferences
    ? 'fetching'
    : 'idle';

  return {
    preferences,
    status,
    isLoading: isUpdatingPreferences,
    isFetching: isFetchingPreferences,
    fetchPreferences,
    updatePreferences
  };
}

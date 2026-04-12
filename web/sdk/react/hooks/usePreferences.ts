import {
  useQuery,
  useMutation,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import {
  UseMutationResult,
  useQueryClient,
  UseQueryResult
} from '@tanstack/react-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  CreateCurrentUserPreferencesRequestSchema,
  ListCurrentUserPreferencesRequestSchema
} from '@raystack/proton/frontier';
import { useCallback, useMemo } from 'react';
import { handleConnectError } from '~/utils/error';

type Preference = {
  name?: string;
  value?: string;
  scopeType?: string;
  scopeId?: string;
  [key: string]: any;
};

type Preferences = Record<string, Preference>;

export interface UsePreferences {
  preferences: Preferences;
  isLoading: boolean;
  isFetching: boolean;
  status: 'idle' | 'fetching' | 'loading';
  fetchPreferences: () => void;
  updatePreferences: (preferences: Preference[]) => Promise<void>;
  fetchPreferencesStatus: UseQueryResult['status'];
  updatePreferencesStatus: UseMutationResult['status'];
}

function getFormattedData(preferences: Preference[] = []): Preferences {
  return preferences.reduce((acc: Preferences, preference) => {
    if (preference?.name) acc[preference.name] = preference;
    return acc;
  }, {});
}

export function usePreferences({
  autoFetch = true,
  scopeType,
  scopeId
}: {
  autoFetch?: boolean;
  scopeType?: string;
  scopeId?: string;
} = {}): UsePreferences {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const {
    data: preferencesData,
    isLoading: isFetchingPreferences,
    refetch,
    status: fetchPreferencesStatus
  } = useQuery(
    FrontierServiceQueries.listCurrentUserPreferences,
    create(ListCurrentUserPreferencesRequestSchema, {
      scopeType,
      scopeId
    }),
    {
      enabled: autoFetch
    }
  );

  const preferences = useMemo(
    () => getFormattedData(preferencesData?.preferences ?? []),
    [preferencesData]
  );

  const {
    mutateAsync: updatePreferencesMutation,
    isPending: isUpdatingPreferences,
    status: updatePreferencesStatus
  } = useMutation(FrontierServiceQueries.createCurrentUserPreferences, {
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.listCurrentUserPreferences,
          transport,
          cardinality: 'finite'
        })
      });
    }
  });

  const updatePreferences = useCallback(
    async (preferences: Preference[]) => {
      try {
        const req = create(CreateCurrentUserPreferencesRequestSchema, {
          bodies: preferences
        });
        await updatePreferencesMutation(req);
      } catch (err) {
        handleConnectError(err, {
          PermissionDenied: () =>
            console.error(
              'frontier:sdk:: Permission denied while updating user preferences'
            ),
          InvalidArgument: e =>
            console.error(
              'frontier:sdk:: Invalid preferences input:',
              e.message
            )
        });
        throw err;
      }
    },
    [updatePreferencesMutation]
  );

  const status: UsePreferences['status'] = isUpdatingPreferences
    ? 'loading'
    : isFetchingPreferences
    ? 'fetching'
    : 'idle';

  return {
    preferences: preferences ?? {},
    status,
    isLoading: isUpdatingPreferences,
    isFetching: isFetchingPreferences,
    fetchPreferences: refetch,
    updatePreferences,
    updatePreferencesStatus,
    fetchPreferencesStatus
  };
}

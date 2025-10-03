import { useEffect } from 'react';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';

// Global state with reference counting (this prevents multiple instances of the hook from causing multiple intervals to be set)
let globalIntervalId: number | null = null;
let trackingCount = 0;

export const useLastActiveTracker = () => {
  const { client } = useFrontier();

  const {
    mutate: pingUserSession,
  } = useMutation(FrontierServiceQueries.pingUserSession, {
    onSuccess: () => {
      console.log('Session pinged successfully');
    },
    onError: (error) => {
      console.error('Failed to ping session:', error);
    },
  });

  useEffect(() => {
    if (!client) return;

    // Increment reference count
    trackingCount++;

    // Start tracking if this is the first component
    if (trackingCount === 1) {
      console.log('Starting activity tracking (first component)');
      pingUserSession({});
      globalIntervalId = setInterval(() => {
        pingUserSession({});
      }, 10 * 60 * 1000);
    } else {
      console.log(`Activity tracking already running (${trackingCount} components)`);
    }

    return () => {
      // Decrement reference count
      trackingCount--;
      console.log(`Component unmounted, tracking count: ${trackingCount}`);
      
      // Stop tracking if this was the last component
      if (trackingCount === 0 && globalIntervalId) {
        console.log('Stopping activity tracking (last component)');
        clearInterval(globalIntervalId);
        globalIntervalId = null;
      }
    };
  }, [client, pingUserSession]);
};
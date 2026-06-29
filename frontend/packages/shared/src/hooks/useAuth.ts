import { useState } from 'react';
import { useAuthContext } from '../providers/AuthProvider';
import type { EnrichedReport } from '../types';

export const useAuth = () => {
  const { login: ctxLogin, logout, user, token, isLoading: ctxLoading } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (phone: string): Promise<EnrichedReport | null> => {
    setLoading(true);
    setError(null);
    try {
      const profile = await ctxLogin(phone);
      return profile;
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Login failed');
      }
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { login, logout, user, token, loading: loading || ctxLoading, error };
};

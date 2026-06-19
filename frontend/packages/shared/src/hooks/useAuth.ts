import { useState } from 'react';
import { useRepositories } from '../providers/RepositoryProvider';
import type { EnrichedReport } from '../types';

export const useAuth = () => {
  const { auth } = useRepositories();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (phone: string): Promise<EnrichedReport | null> => {
    setLoading(true);
    setError(null);
    try {
      const user = await auth.login(phone);
      return user;
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

  return { login, loading, error };
};

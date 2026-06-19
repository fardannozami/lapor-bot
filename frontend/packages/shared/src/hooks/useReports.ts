import { useQuery } from '@tanstack/react-query';
import { useRepositories } from '../providers/RepositoryProvider';
import type { EnrichedReport, GlobalSummary } from '../types';

export const useReports = () => {
  const { reports } = useRepositories();

  const summaryQuery = useQuery<GlobalSummary, Error>({
    queryKey: ['summary'],
    queryFn: () => reports.getSummary(),
  });

  const leaderboardQuery = useQuery<EnrichedReport[], Error>({
    queryKey: ['leaderboard'],
    queryFn: () => reports.getLeaderboard(),
  });

  return {
    summary: summaryQuery.data,
    hunters: leaderboardQuery.data || [],
    loading: summaryQuery.isPending || leaderboardQuery.isPending,
    refreshing: summaryQuery.isRefetching || leaderboardQuery.isRefetching,
    error: summaryQuery.error || leaderboardQuery.error,
    refresh: () => {
      summaryQuery.refetch();
      leaderboardQuery.refetch();
    },
  };
};

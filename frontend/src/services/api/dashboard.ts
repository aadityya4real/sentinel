import { useQuery } from '@tanstack/react-query';
import { apiGet, USE_MOCK_DATA } from '@/services/http';
import { withMockFallback } from '@/services/mock';
import { mockOverview, mockHostsPage } from '@/services/mock/dashboard';
import type { Overview, HostsPage, History } from '@/types/api';

async function fetchOverview(): Promise<Overview> {
  return apiGet<Overview>('/api/v1/dashboard/overview');
}

async function fetchHosts(limit = 100): Promise<HostsPage> {
  return apiGet<HostsPage>('/api/v1/dashboard/hosts', { limit });
}

async function fetchHistory(hostname: string, limit = 300): Promise<History> {
  return apiGet<History>(`/api/v1/dashboard/hosts/${hostname}/metrics`, { limit });
}

function isEmptyOverview(data: Overview): boolean {
  return data.total_hosts === 0;
}

function isEmptyHosts(data: HostsPage): boolean {
  return data.hosts.length === 0;
}

export function useOverview() {
  return useQuery({
    queryKey: ['dashboard', 'overview'],
    queryFn: () =>
      withMockFallback(['dashboard', 'overview'], fetchOverview, mockOverview, {
        enabled: USE_MOCK_DATA,
        isEmpty: isEmptyOverview,
      }),
  });
}

export function useHosts(limit = 100) {
  return useQuery({
    queryKey: ['dashboard', 'hosts', limit],
    queryFn: () =>
      withMockFallback(['dashboard', 'hosts'], () => fetchHosts(limit), mockHostsPage, {
        enabled: USE_MOCK_DATA,
        isEmpty: isEmptyHosts,
      }),
  });
}

export function useHistory(hostname: string, limit = 300) {
  return useQuery({
    queryKey: ['dashboard', 'hosts', hostname, 'history', limit],
    queryFn: () => fetchHistory(hostname, limit),
    enabled: !!hostname,
  });
}

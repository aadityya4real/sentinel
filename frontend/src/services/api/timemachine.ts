import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/services/http';
import type { Snapshot } from '@/types/api';

async function fetchSnapshot(hostname: string, at: string): Promise<Snapshot> {
  return apiGet<Snapshot>(`/api/v1/time-machine/hosts/${hostname}`, { at });
}

export function useSnapshot(hostname: string, at: string) {
  return useQuery({
    queryKey: ['timemachine', hostname, at],
    queryFn: () => fetchSnapshot(hostname, at),
    enabled: !!hostname && !!at,
  });
}

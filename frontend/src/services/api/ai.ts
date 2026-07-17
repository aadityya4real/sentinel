import { useMutation } from '@tanstack/react-query';
import { apiPost, USE_MOCK_DATA } from '@/services/http';
import { withMockMutationFallback } from '@/services/mock';
import { mockAnalysis } from '@/services/mock/ai';
import type { Analysis, IncidentRequest } from '@/types/api';

async function analyzeIncident(req: IncidentRequest): Promise<Analysis> {
  return withMockMutationFallback(
    () => apiPost<Analysis>('/api/v1/ai/incidents/analyze', req),
    mockAnalysis,
    { enabled: USE_MOCK_DATA },
  );
}

export function useAnalyzeIncident() {
  return useMutation({ mutationFn: analyzeIncident });
}

export interface MockConfig<T> {
  enabled: boolean;
  isEmpty: (data: T) => boolean;
}

export async function withMockFallback<T>(
  _queryKey: string[],
  realFetcher: () => Promise<T>,
  mockGenerator: () => T,
  config: MockConfig<T>,
): Promise<T> {
  try {
    const data = await realFetcher();
    if (config.enabled && config.isEmpty(data)) {
      return mockGenerator();
    }
    return data;
  } catch {
    if (config.enabled) {
      return mockGenerator();
    }
    throw new Error('Backend unavailable and mock data is disabled');
  }
}

export async function withMockMutationFallback<T>(
  realFetcher: () => Promise<T>,
  mockGenerator: () => T,
  config: { enabled: boolean },
): Promise<T> {
  try {
    return await realFetcher();
  } catch {
    if (config.enabled) {
      return mockGenerator();
    }
    throw new Error('Backend unavailable and mock data is disabled');
  }
}

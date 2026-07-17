import { useState, useEffect } from 'react';

export function useClock(intervalMs = 1000): string {
  const [time, setTime] = useState(() => new Date());
  useEffect(() => {
    const id = setInterval(() => setTime(new Date()), intervalMs);
    return () => clearInterval(id);
  }, [intervalMs]);
  return time.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

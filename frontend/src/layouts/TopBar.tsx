import { Search, Sun, Moon } from 'lucide-react';
import { useHealth } from '@/services/api/health';
import { useClock } from '@/hooks/useClock';
import { useTheme } from '@/hooks/useTheme';
import { useDebounce } from '@/hooks/useDebounce';
import { useState } from 'react';

export function TopBar() {
  const { data: health, isLoading } = useHealth();
  const clock = useClock();
  const { theme, toggleTheme } = useTheme();
  const [query, setQuery] = useState('');
  useDebounce(query, 300);

  const healthy = health?.status === 'healthy';

  return (
    <header className="sticky top-0 z-20 flex h-16 items-center gap-4 border-b border-line bg-base/80 px-6 backdrop-blur-md">
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search hosts..."
          className="h-9 w-full rounded-lg border border-line bg-surface pl-10 pr-3 text-sm text-slate-200 placeholder:text-slate-500 focus:border-accent focus:outline-none"
          aria-label="Search hosts"
        />
      </div>

      <div className="flex items-center gap-4">
        <div className="hidden items-center gap-2 rounded-lg border border-line bg-surface px-3 py-1.5 sm:flex">
          <span
            className={`h-2 w-2 rounded-full ${isLoading ? 'bg-slate-500 animate-pulse' : healthy ? 'bg-emerald-400' : 'bg-rose-400'}`}
          />
          <span className="text-xs font-medium text-slate-400">
            {isLoading ? 'Checking' : healthy ? 'Operational' : 'Disconnected'}
          </span>
        </div>

        <span className="hidden font-mono text-sm text-slate-400 md:inline" aria-label="Current time">
          {clock}
        </span>

        <button
          onClick={toggleTheme}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-line bg-surface text-slate-400 transition-colors hover:text-slate-200"
          aria-label="Toggle theme"
        >
          {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </button>

        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-accent to-fuchsia-500 text-xs font-semibold text-white">
          A
        </div>
      </div>
    </header>
  );
}

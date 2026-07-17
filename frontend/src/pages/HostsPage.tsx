import { motion } from 'framer-motion';
import { useState } from 'react';
import { useDebounce } from '@/hooks/useDebounce';
import { useHosts } from '@/services/api/dashboard';
import { HostTable } from '@/components/dashboard/HostTable';
import { Search } from 'lucide-react';

export default function HostsPage() {
  const [query, setQuery] = useState('');
  const debounced = useDebounce(query, 200);
  const { data, isLoading, isError, error } = useHosts();

  const filtered = (data?.hosts ?? []).filter((h) =>
    h.metrics.hostname.toLowerCase().includes(debounced.toLowerCase()),
  );

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-100">Hosts</h1>
          <p className="text-sm text-slate-500">{data?.hosts.length ?? 0} hosts in fleet</p>
        </div>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Filter hosts..."
            className="h-9 w-64 rounded-lg border border-line bg-surface pl-10 pr-3 text-sm text-slate-200 placeholder:text-slate-500 focus:border-accent focus:outline-none"
          />
        </div>
      </div>

      {isError ? (
        <div className="card p-5 text-rose-400">{error?.message ?? 'Failed to load hosts'}</div>
      ) : (
        <HostTable hosts={filtered} isLoading={isLoading} />
      )}
    </motion.div>
  );
}

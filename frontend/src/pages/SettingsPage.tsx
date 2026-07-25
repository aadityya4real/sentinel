import { motion } from 'framer-motion';
import { Sun, Moon, Server, Clock, Info } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { useTheme } from '@/hooks/useTheme';
import { useRefreshInterval, REFRESH_INTERVALS, formatRefreshLabel } from '@/hooks/useRefreshInterval';
import { API_URL } from '@/config/env';

interface SettingRowProps {
  label: string;
  children: React.ReactNode;
}

function SettingRow({ label, children }: SettingRowProps) {
  return (
    <div className="flex items-center justify-between py-3">
      <span className="text-sm text-slate-300">{label}</span>
      {children}
    </div>
  );
}

export default function SettingsPage() {
  const { theme, toggleTheme } = useTheme();
  const { ms, setMs, label } = useRefreshInterval();

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="max-w-2xl space-y-6">
      <div>
        <h1 className="text-xl font-bold text-slate-100">Settings</h1>
        <p className="text-sm text-slate-500">Manage your Sentinel preferences</p>
      </div>

      <Card>
        <div className="mb-4 flex items-center gap-2">
          <Sun className="h-4 w-4 text-amber-400" />
          <h2 className="text-sm font-medium text-slate-200">Appearance</h2>
        </div>
        <SettingRow label="Theme">
          <button
            onClick={toggleTheme}
            className="flex h-9 items-center gap-2 rounded-lg border border-line bg-base px-3 text-sm text-slate-300 hover:bg-elevated transition-colors"
          >
            {theme === 'dark' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
            {theme === 'dark' ? 'Dark' : 'Light'}
          </button>
        </SettingRow>
      </Card>

      <Card>
        <div className="mb-4 flex items-center gap-2">
          <Server className="h-4 w-4 text-accent" />
          <h2 className="text-sm font-medium text-slate-200">API</h2>
        </div>
        <SettingRow label="Backend URL">
          <code className="rounded bg-base px-2 py-1 text-xs text-slate-400">
            {API_URL || '/api (proxied)'}
          </code>
        </SettingRow>
        <SettingRow label="Mock data fallback">
          <code className="rounded bg-base px-2 py-1 text-xs text-slate-400">
            {import.meta.env.VITE_USE_MOCK_DATA === 'true' ? 'enabled (dev)' : 'disabled'}
          </code>
        </SettingRow>
      </Card>

      <Card>
        <div className="mb-4 flex items-center gap-2">
          <Clock className="h-4 w-4 text-accent" />
          <h2 className="text-sm font-medium text-slate-200">Data Refresh</h2>
        </div>
        <SettingRow label="Auto-refresh interval">
          <div className="flex items-center gap-2">
            <select
              value={ms}
              onChange={(e) => setMs(Number(e.target.value))}
              className="h-8 rounded-lg border border-line bg-base px-3 text-sm text-slate-200 focus:border-accent focus:outline-none cursor-pointer"
            >
              {REFRESH_INTERVALS.map((interval) => (
                <option key={interval} value={interval}>{formatRefreshLabel(interval)}</option>
              ))}
            </select>
            <span className="text-xs text-slate-500">Current: {label}</span>
          </div>
        </SettingRow>
        <p className="text-xs text-slate-500 pt-1">
          How often the dashboard polls the backend for new metrics. Lower values increase request frequency.
        </p>
      </Card>

      <Card>
        <div className="mb-4 flex items-center gap-2">
          <Info className="h-4 w-4 text-accent" />
          <h2 className="text-sm font-medium text-slate-200">About</h2>
        </div>
        <p className="text-sm text-slate-400">
          Sentinel v{import.meta.env.VITE_APP_VERSION ?? '0.1.0'} — Infrastructure Event Intelligence Platform. Record. Replay. Explain.
        </p>
      </Card>
    </motion.div>
  );
}

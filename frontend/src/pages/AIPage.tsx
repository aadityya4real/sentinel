import { motion } from 'framer-motion';
import { useState, useEffect } from 'react';
import { Brain, Send, AlertCircle, Sparkles } from 'lucide-react';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Spinner } from '@/components/ui/Spinner';
import { useAnalyzeIncident } from '@/services/api/ai';
import type { Analysis } from '@/types/api';

const HISTORY_KEY = 'sentinel-ai-history';

function severityVariant(severity: string): 'info' | 'active' | 'critical' {
  if (severity === 'critical' || severity === 'high') return 'critical';
  if (severity === 'medium') return 'active';
  return 'info';
}

export default function AIPage() {
  const [hostname, setHostname] = useState('prod-api-01');
  const [minutes, setMinutes] = useState(15);
  const [result, setResult] = useState<Analysis | null>(null);
  const [history, setHistory] = useState<Analysis[]>([]);

  const mutation = useAnalyzeIncident();

  useEffect(() => {
    const stored = localStorage.getItem(HISTORY_KEY);
    if (stored) setHistory(JSON.parse(stored));
  }, []);

  function analyze() {
    const now = new Date();
    const from = new Date(now.getTime() - minutes * 60_000);
    mutation.mutate(
      { hostname, from: from.toISOString(), to: now.toISOString(), event_limit: 50 },
      {
        onSuccess: (data) => {
          setResult(data);
          const updated = [data, ...history].slice(0, 10);
          setHistory(updated);
          localStorage.setItem(HISTORY_KEY, JSON.stringify(updated));
        },
      },
    );
  }

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <div className="lg:col-span-1">
        <Card>
          <div className="mb-4 flex items-center gap-3">
            <div className="rounded-xl bg-violet-500/10 p-2.5">
              <Brain className="h-5 w-5 text-violet-400" />
            </div>
            <h2 className="text-sm font-medium text-slate-200">Incident Analysis</h2>
          </div>
          <div className="space-y-4">
            <div>
              <label className="mb-1.5 block text-xs font-medium text-slate-500">Hostname</label>
              <input
                value={hostname}
                onChange={(e) => setHostname(e.target.value)}
                className="h-9 w-full rounded-lg border border-line bg-base px-3 text-sm text-slate-200 focus:border-accent focus:outline-none"
                placeholder="e.g. prod-api-01"
              />
            </div>
            <div>
              <label className="mb-1.5 block text-xs font-medium text-slate-500">Window (minutes)</label>
              <input
                type="number"
                value={minutes}
                onChange={(e) => setMinutes(Number(e.target.value))}
                min={1}
                max={1440}
                className="h-9 w-full rounded-lg border border-line bg-base px-3 text-sm text-slate-200 focus:border-accent focus:outline-none"
              />
            </div>
            <Button onClick={analyze} disabled={mutation.isPending || !hostname} className="w-full">
              {mutation.isPending ? <Spinner size="sm" /> : <Send className="h-4 w-4" />}
              Analyze Incident
            </Button>
          </div>
        </Card>

        {history.length > 0 && (
          <Card className="mt-4">
            <h3 className="mb-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Recent Analyses</h3>
            <div className="space-y-2">
              {history.map((h, i) => (
                <button
                  key={i}
                  onClick={() => setResult(h)}
                  className="w-full rounded-lg p-2 text-left transition-colors hover:bg-elevated"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-slate-300">{h.hostname}</span>
                    <Badge variant={severityVariant(h.severity)}>{h.severity}</Badge>
                  </div>
                  <p className="mt-0.5 truncate text-xs text-slate-500">{h.summary}</p>
                </button>
              ))}
            </div>
          </Card>
        )}
      </div>

      <div className="lg:col-span-2">
        {mutation.isPending ? (
          <Card className="flex h-full items-center justify-center py-24">
            <div className="text-center">
              <Sparkles className="mx-auto mb-4 h-10 w-10 text-violet-400 animate-pulse" />
              <p className="text-sm text-slate-400">Analyzing incident...</p>
            </div>
          </Card>
        ) : mutation.isError ? (
          <Card className="flex h-full items-center justify-center py-24">
            <div className="text-center">
              <AlertCircle className="mx-auto mb-4 h-10 w-10 text-rose-400" />
              <p className="text-sm text-slate-400">{(mutation.error as Error)?.message ?? 'Analysis failed'}</p>
            </div>
          </Card>
        ) : result ? (
          <Card>
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h2 className="text-sm font-medium text-slate-200">{result.hostname}</h2>
                <p className="text-xs text-slate-500">
                  {new Date(result.from).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })} –{' '}
                  {new Date(result.to).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={severityVariant(result.severity)}>{result.severity}</Badge>
                <span className="text-xs text-slate-500">{(result.confidence * 100).toFixed(0)}% confidence</span>
              </div>
            </div>

            <div className="mb-2 h-1.5 w-full overflow-hidden rounded-full bg-line">
              <div className="h-full rounded-full bg-accent" style={{ width: `${result.confidence * 100}%` }} />
            </div>

            <p className="mb-6 text-sm leading-relaxed text-slate-300">{result.summary}</p>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <h4 className="mb-2 text-xs font-medium text-slate-500 uppercase tracking-wider">Probable Causes</h4>
                <ul className="space-y-1.5">
                  {result.probable_causes.map((cause, i) => (
                    <li key={i} className="flex gap-2 text-sm text-slate-400">
                      <span className="text-accent">•</span> {cause}
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <h4 className="mb-2 text-xs font-medium text-slate-500 uppercase tracking-wider">Recommended Actions</h4>
                <ul className="space-y-1.5">
                  {result.recommended_actions.map((action, i) => (
                    <li key={i} className="flex gap-2 text-sm text-slate-400">
                      <span className="text-emerald-400">→</span> {action}
                    </li>
                  ))}
                </ul>
              </div>
            </div>

            {result.evidence.length > 0 && (
              <div className="mt-6">
                <h4 className="mb-2 text-xs font-medium text-slate-500 uppercase tracking-wider">Evidence</h4>
                <div className="space-y-2">
                  {result.evidence.map((ev) => (
                    <div key={ev.event_id} className="rounded-lg bg-elevated p-3">
                      <span className="font-mono text-xs text-accent-bright">#{ev.event_id}</span>
                      <p className="mt-1 text-sm text-slate-300">{ev.observation}</p>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </Card>
        ) : (
          <Card className="flex h-full items-center justify-center py-24">
            <div className="text-center">
              <Brain className="mx-auto mb-4 h-10 w-10 text-slate-600" />
              <p className="text-sm text-slate-400">Submit an incident to generate an AI analysis.</p>
            </div>
          </Card>
        )}
      </div>
    </motion.div>
  );
}

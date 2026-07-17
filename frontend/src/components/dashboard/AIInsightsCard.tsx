import { Brain, Sparkles } from 'lucide-react';
import { Card } from '@/components/ui/Card';

export function AIInsightsCard() {
  return (
    <Card>
      <div className="flex items-center gap-3 mb-4">
        <div className="rounded-xl bg-violet-500/10 p-2.5">
          <Brain className="h-5 w-5 text-violet-400" />
        </div>
        <div>
          <h3 className="text-sm font-medium text-slate-200">AI Insights</h3>
          <p className="text-xs text-slate-500">Powered by incident analysis</p>
        </div>
      </div>
      <div className="rounded-xl bg-elevated p-6 text-center">
        <Sparkles className="mx-auto mb-3 h-8 w-8 text-violet-500/40" />
        <p className="text-sm text-slate-400">AI analysis requires an active backend connection.</p>
        <p className="mt-1 text-xs text-slate-500">Navigate to the AI page to analyze incidents.</p>
      </div>
    </Card>
  );
}

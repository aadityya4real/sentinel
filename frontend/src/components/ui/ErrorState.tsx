import { Button } from './Button';
import { AlertTriangle } from 'lucide-react';

interface ErrorStateProps {
  title?: string;
  message?: string;
  onRetry?: () => void;
}

export function ErrorState({ title = 'Something went wrong', message, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 rounded-2xl bg-rose-500/10 p-4">
        <AlertTriangle className="h-8 w-8 text-rose-400" />
      </div>
      <h3 className="text-sm font-medium text-slate-300">{title}</h3>
      {message && <p className="mt-1 max-w-xs text-xs text-slate-500">{message}</p>}
      {onRetry && (
        <Button variant="secondary" size="sm" className="mt-4" onClick={onRetry}>
          Retry
        </Button>
      )}
    </div>
  );
}

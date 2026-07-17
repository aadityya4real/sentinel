import type { LucideIcon } from 'lucide-react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
}

export function EmptyState({ icon: Icon = Inbox, title, description }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 rounded-2xl bg-elevated p-4">
        <Icon className="h-8 w-8 text-slate-500" />
      </div>
      <h3 className="text-sm font-medium text-slate-300">{title}</h3>
      {description && <p className="mt-1 text-xs text-slate-500">{description}</p>}
    </div>
  );
}

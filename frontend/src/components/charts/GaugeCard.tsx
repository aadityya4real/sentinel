import { Card } from '@/components/ui/Card';

interface GaugeCardProps {
  title: string;
  value: number;
  max?: number;
  color?: string;
  unit?: string;
}

export function GaugeCard({ title, value, max = 100, color = '#7c3aed', unit = '%' }: GaugeCardProps) {
  const radius = 52;
  const circumference = 2 * Math.PI * radius;
  const pct = Math.min(value / max, 1);
  const offset = circumference - pct * circumference;

  return (
    <Card className="flex flex-col items-center">
      <h3 className="mb-4 self-start text-sm font-medium text-slate-200">{title}</h3>
      <div className="relative h-32 w-32">
        <svg className="h-full w-full -rotate-90" viewBox="0 0 120 120">
          <circle cx="60" cy="60" r={radius} fill="none" stroke="#27272f" strokeWidth="10" />
          <circle
            cx="60"
            cy="60"
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth="10"
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            style={{ transition: 'stroke-dashoffset 0.5s ease' }}
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-2xl font-bold text-slate-100">{value.toFixed(0)}</span>
          <span className="text-xs text-slate-500">{unit}</span>
        </div>
      </div>
    </Card>
  );
}

import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { Card } from '@/components/ui/Card';

interface AreaChartCardProps {
  title: string;
  data: { timestamp: string; value: number }[];
  color?: string;
  unit?: string;
}

interface TooltipPayloadItem {
  value: number;
}

function ChartTooltip({ active, payload, label, unit }: { active?: boolean; payload?: TooltipPayloadItem[]; label?: string; unit: string }) {
  if (!active || !payload || payload.length === 0) return null;
  return (
    <div className="rounded-lg border border-line bg-elevated px-3 py-2 shadow-card">
      <p className="text-xs text-slate-400">{label}</p>
      <p className="text-sm font-medium text-slate-100">
        {payload[0].value.toFixed(1)}
        {unit}
      </p>
    </div>
  );
}

export function AreaChartCard({ title, data, color = '#7c3aed', unit = '%' }: AreaChartCardProps) {
  const gradientId = `gradient-${title.replace(/\s+/g, '-')}`;
  return (
    <Card>
      <h3 className="mb-4 text-sm font-medium text-slate-200">{title}</h3>
      <div className="h-48">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 4, right: 4, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.4} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis
              dataKey="timestamp"
              tick={{ fill: '#64748b', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fill: '#64748b', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              domain={[0, 100]}
            />
            <Tooltip content={<ChartTooltip unit={unit} />} />
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              fill={`url(#${gradientId})`}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </Card>
  );
}

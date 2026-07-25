import { motion } from 'framer-motion';
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid } from 'recharts';

interface InfChartProps {
  title: string;
  data: { timestamp: string; value: number }[];
  color: string;
  unit: string;
  isLoading?: boolean;
}

interface ChartTooltipProps {
  active?: boolean;
  payload?: Array<{ value: number }>;
  label?: string;
  unit: string;
}

function ChartTooltip({ active, payload, label, unit }: ChartTooltipProps) {
  if (!active || !payload?.length) return null;
  return (
    <div className="rounded-lg border border-line bg-surface px-3 py-2 shadow-card">
      <p className="text-xs text-text-muted">{label}</p>
      <p className="text-sm font-medium text-text-primary">
        {payload[0].value.toFixed(1)}{unit}
      </p>
    </div>
  );
}

function yDomain(data: { value: number }[]) {
  const maxVal = data.length ? Math.max(...data.map((d) => d.value)) : 100;
  return [0, Math.min(100, Math.ceil(maxVal / 10) * 10 + 10)] as [number, number];
}

export function InfChart({ title, data, color, unit, isLoading }: InfChartProps) {
  const gradientId = `chart-${title.replace(/\s+/g, '_')}`;
  const domain = yDomain(data);

  if (isLoading || data.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 4 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="card flex items-center justify-center p-8"
      >
        <div className="text-center">
          <div className="mx-auto mb-2 h-8 w-8 rounded-full bg-elevated animate-shimmer" style={{ backgroundImage: 'linear-gradient(90deg, transparent, rgba(139,92,246,0.06), transparent)', backgroundSize: '200% 100%' }} />
          <p className="text-xs text-text-muted">
            {isLoading ? 'Loading...' : 'No data yet'}
          </p>
        </div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="card"
    >
      <h3 className="mb-3 text-sm font-medium text-text-primary">{title}</h3>
      <div className="h-44">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 4, right: 4, left: -16, bottom: 0 }}>
            <defs>
              <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.35} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border-line)" vertical={false} />
            <XAxis
              dataKey="timestamp"
              tick={{ fill: 'var(--text-muted)', fontSize: 10 }}
              axisLine={false}
              tickLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fill: 'var(--text-muted)', fontSize: 10 }}
              axisLine={false}
              tickLine={false}
              domain={domain}
              tickCount={5}
            />
            <Tooltip content={<ChartTooltip unit={unit} />} />
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              fill={`url(#${gradientId})`}
              isAnimationActive
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </motion.div>
  );
}

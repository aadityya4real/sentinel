import { motion } from 'framer-motion';
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis, CartesianGrid } from 'recharts';

/**
 * Infrastructure chart — dark theme, gradient fill, smooth interpolation.
 * Used for CPU, memory, disk and network panels on the dashboard.
 */
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
    <div className="rounded-lg border border-line bg-elevated px-3 py-2 shadow-card">
      <p className="text-xs text-slate-400">{label}</p>
      <p className="text-sm font-medium text-slate-100">
        {payload[0].value.toFixed(1)}{unit}
      </p>
    </div>
  );
}

function yDomain(data: { value: number }[]): [number, number] {
  const maxVal = data.length ? Math.max(...data.map((d) => d.value)) : 100;
  return [0, Math.min(100, Math.ceil(maxVal / 10) * 10 + 10)];
}

export function InfChart({ title, data, color, unit, isLoading }: InfChartProps) {
  const gradientId = `g-${title.replace(/\s+/g, '-')}`;
  const domain = yDomain(data);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="card"
    >
      <h3 className="mb-3 text-sm font-medium text-slate-200">{title}</h3>
      <div className="h-44">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={isLoading ? [] : data} margin={{ top: 4, right: 4, left: -16, bottom: 0 }}>
            <defs>
              <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.35} />
                <stop offset="100%" stopColor={color} stopOpacity={0} />
              </linearGradient>
            </defs>
            {data.length > 0 && (
              <>
                <CartesianGrid strokeDasharray="3 3" stroke="#1e1e2a" vertical={false} />
                <XAxis
                  dataKey="timestamp"
                  tick={{ fill: '#475569', fontSize: 10 }}
                  axisLine={false}
                  tickLine={false}
                  interval="preserveStartEnd"
                />
                <YAxis
                  tick={{ fill: '#475569', fontSize: 10 }}
                  axisLine={false}
                  tickLine={false}
                  domain={domain}
                  tickCount={5}
                />
              </>
            )}
            <Tooltip content={<ChartTooltip unit={unit} />} />
            <Area
              type="monotone"
              dataKey="value"
              stroke={color}
              strokeWidth={2}
              fill={`url(#${gradientId})`}
              isAnimationActive={!isLoading}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </motion.div>
  );
}

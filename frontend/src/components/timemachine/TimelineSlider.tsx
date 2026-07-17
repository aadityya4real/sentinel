interface TimelineSliderProps {
  min: number;
  max: number;
  value: number;
  onChange: (value: number) => void;
}

function formatTick(ts: number): string {
  return new Date(ts).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
}

export function TimelineSlider({ min, max, value, onChange }: TimelineSliderProps) {
  const range = max - min;
  const pct = range > 0 ? ((value - min) / range) * 100 : 0;

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-xs text-slate-500">
        <span>{new Date(min).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })}</span>
        <span>{new Date(max).toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' })}</span>
      </div>
      <div className="relative">
        <input
          type="range"
          min={min}
          max={max}
          value={value}
          step={60000}
          onChange={(e) => onChange(Number(e.target.value))}
          className="h-2 w-full cursor-pointer appearance-none rounded-full bg-line outline-none accent-accent"
          style={{
            background: `linear-gradient(to right, #7c3aed ${pct}%, #27272f ${pct}%)`,
          }}
          aria-label="Timeline position"
        />
      </div>
      <p className="text-center font-mono text-sm text-accent-bright">{formatTick(value)}</p>
    </div>
  );
}

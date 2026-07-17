import { SkipBack, Play, Pause, SkipForward } from 'lucide-react';

interface ReplayControlsProps {
  isPlaying: boolean;
  onPlay: () => void;
  onPause: () => void;
  onPrev: () => void;
  onNext: () => void;
  canPrev: boolean;
  canNext: boolean;
}

export function ReplayControls({ isPlaying, onPlay, onPause, onPrev, onNext, canPrev, canNext }: ReplayControlsProps) {
  return (
    <div className="flex items-center justify-center gap-3">
      <button
        onClick={onPrev}
        disabled={!canPrev}
        className="flex h-10 w-10 items-center justify-center rounded-lg border border-line bg-surface text-slate-300 transition-colors hover:bg-elevated disabled:opacity-30"
        aria-label="Previous"
      >
        <SkipBack className="h-4 w-4" />
      </button>

      {isPlaying ? (
        <button
          onClick={onPause}
          className="flex h-12 w-12 items-center justify-center rounded-full bg-accent text-white shadow-glow transition-transform hover:scale-105"
          aria-label="Pause"
        >
          <Pause className="h-5 w-5" />
        </button>
      ) : (
        <button
          onClick={onPlay}
          className="flex h-12 w-12 items-center justify-center rounded-full bg-accent text-white shadow-glow transition-transform hover:scale-105"
          aria-label="Play"
        >
          <Play className="ml-0.5 h-5 w-5" />
        </button>
      )}

      <button
        onClick={onNext}
        disabled={!canNext}
        className="flex h-10 w-10 items-center justify-center rounded-lg border border-line bg-surface text-slate-300 transition-colors hover:bg-elevated disabled:opacity-30"
        aria-label="Next"
      >
        <SkipForward className="h-4 w-4" />
      </button>
    </div>
  );
}

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        base: 'var(--bg-base, #0a0a0f)',
        surface: 'var(--bg-surface, #13131a)',
        elevated: 'var(--bg-elevated, #1c1c26)',
        line: 'var(--border-line, #27272f)',
        accent: {
          DEFAULT: 'var(--accent, #7c3aed)',
          bright: 'var(--accent-bright, #8b5cf6)',
          soft: 'rgba(124, 58, 237, 0.2)',
        },
        text: {
          primary: 'var(--text-primary, #f1f5f9)',
          secondary: 'var(--text-secondary, #94a3b8)',
          muted: 'var(--text-muted, #475569)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        card: '0 1px 3px 0 rgb(0 0 0 / 0.3), 0 1px 2px -1px rgb(0 0 0 / 0.3)',
        glow: '0 0 24px -6px var(--glow-shadow, rgba(124, 58, 237, 0.45))',
      },
      keyframes: {
        'fade-in': {
          from: { opacity: '0', transform: 'translateY(4px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        shimmer: {
          '100%': { transform: 'translateX(100%)' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.3s ease-out',
        shimmer: 'shimmer 1.5s infinite',
      },
    },
  },
  darkMode: 'class',
};

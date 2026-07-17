import { motion } from 'framer-motion';
import { cn } from '@/lib/cn';
import type { ReactNode } from 'react';

interface CardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  padding?: boolean;
}

export function Card({ children, className, hover = false, padding = true }: CardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.25 }}
      whileHover={hover ? { y: -2, transition: { duration: 0.15 } } : undefined}
      className={cn(
        'card',
        padding && 'p-5',
        hover && 'cursor-pointer transition-shadow hover:shadow-glow',
        className,
      )}
    >
      {children}
    </motion.div>
  );
}

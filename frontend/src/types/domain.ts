import type { LucideIcon } from 'lucide-react';

export interface NavItem {
  label: string;
  path: string;
  icon: LucideIcon;
}

export type HostStatus = 'active' | 'stale';

export type Severity = 'info' | 'low' | 'medium' | 'high' | 'critical';

export type Theme = 'dark' | 'light';

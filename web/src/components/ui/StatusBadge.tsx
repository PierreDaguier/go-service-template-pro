import clsx from 'clsx';

interface StatusBadgeProps {
  tone: 'healthy' | 'degraded' | 'warning' | 'neutral' | 'error';
  label: string;
}

const toneClass: Record<StatusBadgeProps['tone'], string> = {
  healthy: 'badge badge-healthy',
  degraded: 'badge badge-degraded',
  warning: 'badge badge-warning',
  neutral: 'badge badge-neutral',
  error: 'badge badge-error',
};

export function StatusBadge({ tone, label }: StatusBadgeProps) {
  return <span className={clsx(toneClass[tone])}>{label}</span>;
}

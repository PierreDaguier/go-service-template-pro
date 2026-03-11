interface MetricCardProps {
  title: string;
  value: string;
  subtitle?: string;
  trend?: string;
}

export function MetricCard({ title, value, subtitle, trend }: MetricCardProps) {
  return (
    <article className="metric-card">
      <p className="metric-card__title">{title}</p>
      <p className="metric-card__value">{value}</p>
      <div className="metric-card__meta">
        {subtitle ? <span>{subtitle}</span> : <span>&nbsp;</span>}
        {trend ? <span className="metric-card__trend">{trend}</span> : null}
      </div>
    </article>
  );
}

interface StatePanelProps {
  title: string;
  message: string;
  variant: 'loading' | 'error' | 'empty' | 'warning';
}

export function StatePanel({ title, message, variant }: StatePanelProps) {
  return (
    <section className={`state-panel state-panel--${variant}`}>
      <h3>{title}</h3>
      <p>{message}</p>
    </section>
  );
}

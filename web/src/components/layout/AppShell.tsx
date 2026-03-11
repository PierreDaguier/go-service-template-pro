import { Activity, AlertTriangle, Gauge, Settings2, ShieldCheck } from 'lucide-react';
import { NavLink, Outlet } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Service Overview', icon: Gauge, end: true },
  { to: '/metrics', label: 'Live Metrics', icon: Activity },
  { to: '/errors', label: 'Error Explorer', icon: AlertTriangle },
  { to: '/traces', label: 'Trace Explorer', icon: ShieldCheck },
  { to: '/config', label: 'Config & Environment', icon: Settings2 },
];

export function AppShell() {
  return (
    <div className="shell">
      <aside className="shell__sidebar">
        <div className="brand">
          <p className="brand__eyebrow">go-service-template-pro</p>
          <h1>Operations Control Panel</h1>
          <p>Client-facing observability surface for delivery confidence.</p>
        </div>
        <nav className="nav">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                className={({ isActive }) => (isActive ? 'nav__link nav__link--active' : 'nav__link')}
                key={item.to}
                to={item.to}
                end={item.end}
              >
                <Icon size={16} />
                <span>{item.label}</span>
              </NavLink>
            );
          })}
        </nav>
      </aside>

      <main className="shell__content">
        <Outlet />
      </main>
    </div>
  );
}

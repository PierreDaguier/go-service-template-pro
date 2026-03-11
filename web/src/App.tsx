import { Navigate, Route, Routes } from 'react-router-dom';

import { AppShell } from './components/layout/AppShell';
import { ConfigPage } from './pages/ConfigPage';
import { ErrorExplorerPage } from './pages/ErrorExplorerPage';
import { LiveMetricsPage } from './pages/LiveMetricsPage';
import { OverviewPage } from './pages/OverviewPage';
import { TraceExplorerPage } from './pages/TraceExplorerPage';

export function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<OverviewPage />} />
        <Route path="metrics" element={<LiveMetricsPage />} />
        <Route path="errors" element={<ErrorExplorerPage />} />
        <Route path="traces" element={<TraceExplorerPage />} />
        <Route path="config" element={<ConfigPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}

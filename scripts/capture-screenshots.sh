#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

cd "$ROOT_DIR/web"
npm install
VITE_USE_MOCKS=true npm run build
npm install --no-save playwright
npx playwright install chromium

npm run preview >/tmp/ops-preview.log 2>&1 &
PREVIEW_PID=$!
trap 'kill $PREVIEW_PID >/dev/null 2>&1 || true' EXIT

sleep 3

node <<'NODE'
const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1600, height: 980 } });
  const captures = [
    { navLabel: null, file: 'overview.png' },
    { navLabel: 'Live Metrics', file: 'live-metrics.png' },
    { navLabel: 'Error Explorer', file: 'error-explorer.png' },
    { navLabel: 'Trace Explorer', file: 'trace-explorer.png' },
    { navLabel: 'Config & Environment', file: 'config-status.png' }
  ];

  const outDir = path.resolve(__dirname, '..', 'assets', 'screenshots');
  fs.mkdirSync(outDir, { recursive: true });

  await page.goto('http://localhost:4173/', { waitUntil: 'networkidle' });
  await page.waitForTimeout(1200);

  for (const item of captures) {
    if (item.navLabel) {
      await page.getByRole('link', { name: item.navLabel }).click();
      await page.waitForTimeout(1200);
    }
    await page.screenshot({ path: path.join(outDir, item.file), fullPage: true });
  }

  await browser.close();
})();
NODE

kill $PREVIEW_PID >/dev/null 2>&1 || true
trap - EXIT

echo "Screenshots saved to assets/screenshots"

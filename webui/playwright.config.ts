import { defineConfig, devices } from '@playwright/test'

// End-to-end tests of the production build against a mocked discovery API; see
// playwright.live.config.ts for the tests against a real Ghostwire service.
// Additionally, tests a production build with non-default themes and branding.

const port = 4173
const brandedPort = 4174
const brandedOutDir = 'build-branded'

export default defineConfig({
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
    use: {
        baseURL: `http://127.0.0.1:${port}`,
        trace: 'retain-on-failure',
    },
    projects: [
        {
            name: 'chromium',
            testDir: './e2e/mocked',
            use: { ...devices['Desktop Chrome'] },
        },
        {
            name: 'branded',
            testDir: './e2e/branded',
            use: { ...devices['Desktop Chrome'], baseURL: `http://127.0.0.1:${brandedPort}` },
        },
    ],
    webServer: [
        {
            command: `npx vite build && npx vite preview --host 127.0.0.1 --port ${port} --strictPort`,
            url: `http://127.0.0.1:${port}`,
            reuseExistingServer: false,
            timeout: 180_000,
        },
        {
            command: `npx vite build --outDir ${brandedOutDir} && npx vite preview --outDir ${brandedOutDir} --host 127.0.0.1 --port ${brandedPort} --strictPort`,
            env: { VITE_APP_CONFIG: 'e2e/branded/overlay/config.tsx' },
            url: `http://127.0.0.1:${brandedPort}`,
            reuseExistingServer: false,
            timeout: 180_000,
        },
    ],
})

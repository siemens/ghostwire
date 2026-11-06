import { defineConfig, devices } from '@playwright/test'

// End-to-end tests of the production build against a mocked discovery API; see
// playwright.live.config.ts for the tests against a real Ghostwire service.

const port = 4173

export default defineConfig({
    testDir: './e2e/mocked',
    fullyParallel: true,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
    use: {
        baseURL: `http://127.0.0.1:${port}`,
        trace: 'retain-on-failure',
    },
    projects: [
        { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    ],
    webServer: {
        command: `npx vite build && npx vite preview --host 127.0.0.1 --port ${port} --strictPort`,
        url: `http://127.0.0.1:${port}`,
        reuseExistingServer: false,
        timeout: 180_000,
    },
})

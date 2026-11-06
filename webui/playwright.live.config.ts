import { defineConfig, devices } from '@playwright/test'

import { baseURL } from './e2e/live/setup'

// End-to-end tests against a real containerized Ghostwire service, together
// with some well-known containers; see e2e/live/setup.ts.

export default defineConfig({
    testDir: './e2e/live',
    globalSetup: './e2e/live/setup.ts',
    workers: 1,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
    expect: { timeout: 15_000 },
    use: {
        baseURL,
        trace: 'retain-on-failure',
    },
    projects: [
        { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    ],
})

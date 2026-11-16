import { expectChapterRendered, gotoAndDiscover, main, openDrawer } from '../fixtures'
import { discovery, expect, netnsCard, netnsIdOf, test } from './discovery'

// Production builds of React only report minified error messages.
const reactNotAChild = /Minified React error #31/

test.describe('error handling', () => {

    test.describe('crashing view', () => {

        test.use({
            discoveryData: {
                ...discovery,
                metadata: { ...discovery.metadata, hostname: { malformed: true } },
            },
            allowedConsoleErrors: [reactNotAChild],
        })

        test('keeps the app usable and recovers on navigation', async ({ page }) => {
            await gotoAndDiscover(page, '/n')
            await expect(main(page)).toContainText("This view couldn't be shown")
            await expect(page.getByRole('banner')).toBeVisible()

            await openDrawer(page)
            await page.getByRole('link', { name: 'Settings', exact: true }).click()
            await expect(page).toHaveURL(/\/settings$/)
            await expect(main(page)).not.toContainText("couldn't be shown")
        })

    })

    test.describe('crashing network namespace card', () => {

        const crashingNetnsId = netnsIdOf('mock_bar_1')

        const malformed = structuredClone(discovery)
        const crashingNetns = malformed['network-namespaces']
            .find((netns: { netnsid: number }) => netns.netnsid === crashingNetnsId)
        crashingNetns['transport-ports'].ipv4[0]['local-servicename'] = { malformed: true }

        test.use({
            discoveryData: malformed,
            allowedConsoleErrors: [reactNotAChild],
        })

        test('only replaces the crashing card', async ({ page }) => {
            await gotoAndDiscover(page, '/n')
            await expect(main(page)).toContainText(
                `network namespace ${crashingNetnsId} couldn't be shown`)
            await expect(netnsCard(page, crashingNetnsId)).toHaveCount(0)
            await expect(netnsCard(page, netnsIdOf('mock_foo_1'))).toBeAttached()
        })

    })

    test.describe('help chapter failing to load', () => {

        test.use({
            allowedConsoleErrors: [
                /Failed to load resource|Failed to fetch dynamically imported module/,
            ],
        })

        const chapterChunk = /\/assets\/Refresh-[^/]+\.js$/

        test('offers to retry', async ({ page }) => {
            let failing = true
            await page.route(chapterChunk, route => failing ? route.abort() : route.fallback())

            await page.goto('/help/refresh')
            const dialog = page.getByRole('dialog')
            await expect(dialog).toContainText("Content couldn't be loaded")
            await dialog.getByRole('button', { name: 'Close' }).click()
            await expect(dialog).toBeHidden()
            await expect(main(page)).toContainText("This content couldn't be loaded")

            failing = false
            await main(page).getByRole('button', { name: 'Retry' }).click()
            await expectChapterRendered(page)
            await expect(main(page)).not.toContainText("couldn't be loaded")
        })

        test('offers to reload the page', async ({ page }) => {
            await page.route(chapterChunk, route => route.abort(), { times: 1 })

            await page.goto('/help/refresh')
            const dialog = page.getByRole('dialog')
            await expect(dialog).toBeVisible()
            const reloaded = page.waitForEvent('load')
            await dialog.getByRole('button', { name: 'Reload page' }).click()
            await reloaded
            await expectChapterRendered(page)
            await expect(dialog).toBeHidden()
        })

    })

    test.describe('unhandled errors', () => {

        test.use({ allowedConsoleErrors: [/boom/] })

        test('are posted to the snackbar', async ({ page }) => {
            await page.goto('/w')
            await expect(main(page)).toContainText('nothing discovered yet')

            await page.evaluate(() => { setTimeout(() => { throw new Error('boom: handler') }) })
            await expect(page.getByText('unexpected error: boom: handler')).toBeVisible()

            await page.evaluate(() => { void Promise.reject(new Error('boom: rejection')) })
            await expect(page.getByText('unexpected error: boom: rejection')).toBeVisible()
        })

    })

})

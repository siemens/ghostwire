import { discover, gotoAndDiscover, main } from '../fixtures'
import { expect, shownNetnsIds, netnsCard, test } from './discovery'

test.describe('discovery', () => {

    test('does not discover until the user refreshes', async ({ page }) => {
        const discoveries: string[] = []
        page.on('request', req => {
            if (new URL(req.url()).pathname === '/json') {
                discoveries.push(req.url())
            }
        })
        await page.goto('/w')
        await page.waitForLoadState('networkidle')
        expect(discoveries).toHaveLength(0)
        await expect(main(page)).toContainText('nothing discovered yet')

        await discover(page)
        await expect(netnsCard(page, shownNetnsIds[0])).toBeVisible()
        expect(discoveries).toHaveLength(1)
    })

    test.describe('failing', () => {

        test.use({
            discoveryStatus: 500,
            allowedConsoleErrors: [/the server responded with a status of 500/],
        })

        test('notifies about the failure and shows no namespaces', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            await expect(page.getByText(/^refreshing failed: 500/).first()).toBeVisible()
            await expect(main(page)).toContainText('nothing discovered yet')
        })

    })

})

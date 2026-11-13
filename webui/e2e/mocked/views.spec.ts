import { gotoAndDiscover, main, openDrawer } from '../fixtures'
import {
    discovery, expect, hiddenNetnsIds, netnsCard, netnsIdOf, shownNetnsIds, test,
} from './discovery'

test.describe('views', () => {

    test('redirect the root to the wiring view', async ({ page }) => {
        await page.goto('/')
        await expect(page).toHaveURL(/\/w$/)
    })

    test.describe('wiring', () => {

        test('shows all but the "lo"-only process network namespaces', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            for (const id of shownNetnsIds) {
                await expect(netnsCard(page, id)).toBeVisible()
            }
            for (const id of hiddenNetnsIds) {
                await expect(netnsCard(page, id)).toHaveCount(0)
            }
        })

        test('zooms into the network namespace of a containee', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            await main(page).getByRole('button').filter({ hasText: 'mock_bar_1' }).click()
            await expect(page).toHaveURL(/\/w\/mock_bar_1$/)
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toBeAttached()
            await expect(netnsCard(page, netnsIdOf('mock_foo_1'))).toHaveCount(0)
            await main(page).getByRole('link', { name: 'back to overall view' }).click()
            await expect(page).toHaveURL(/\/w$/)
        })

    })

    test.describe('details', () => {

        test('shows all but the "lo"-only process network namespaces', async ({ page }) => {
            await gotoAndDiscover(page, '/n')
            for (const id of shownNetnsIds) {
                await expect(netnsCard(page, id)).toBeAttached()
            }
            for (const id of hiddenNetnsIds) {
                await expect(netnsCard(page, id)).toHaveCount(0)
            }
        })

        test('maximizes and minimizes a single network namespace', async ({ page }) => {
            const id = netnsIdOf('mock_bar_1')
            await gotoAndDiscover(page, '/n')
            await netnsCard(page, id).locator('..')
                .getByRole('link', { name: 'show only this network namespace' }).click()
            await expect(page).toHaveURL(new RegExp(`/n/${id}$`))
            await expect(netnsCard(page, id)).toBeAttached()
            await expect(netnsCard(page, netnsIdOf('mock_foo_1'))).toHaveCount(0)
            await main(page).getByRole('link', { name: 'back to overall view' }).click()
            await expect(page).toHaveURL(/\/n$/)
        })

        test('deep-links to the network namespace of a containee name', async ({ page }) => {
            await gotoAndDiscover(page, '/n/mock_foo_1')
            await expect(netnsCard(page, netnsIdOf('mock_foo_1'))).toBeAttached()
        })

        test('suggests similar containees for an unknown name', async ({ page }) => {
            await gotoAndDiscover(page, '/n/mockbar')
            await expect(main(page)).toContainText('unknown network namespace identifier')
            await expect(main(page)).toContainText('Did you mean')
            await main(page).getByRole('button').filter({ hasText: 'mock_bar_1' }).click()
            await expect(page).toHaveURL(/\/n\/mock_bar_1$/)
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toBeAttached()
        })

    })

    test('lists open and forwarding host ports', async ({ page }) => {
        await gotoAndDiscover(page, '/lochla')
        await expect(main(page)).toContainText('Open & Forwarding Host Ports')
        await expect(main(page).getByRole('cell', { name: 'ssh', exact: true }).first()).toBeVisible()
        await expect(main(page).getByRole('cell', { name: 'ipp', exact: true }).first()).toBeVisible()
    })

    test('shows app and discovery engine versions', async ({ page }) => {
        await gotoAndDiscover(page, '/about')
        await expect(main(page)).toContainText(/web app version:\s*\S+/)
        await expect(main(page)).toContainText(
            `discovery backend engine: ${discovery.metadata['creator-id']} ${discovery.metadata['creator-version']}`)
    })

    test.describe('navigation drawer', () => {

        test('links to all views', async ({ page }) => {
            const views: [string, RegExp][] = [
                ['Open & Forwarding Host Ports', /\/lochla$/],
                ['Network Namespace Details', /\/n$/],
                ['Settings', /\/settings$/],
                ['Help', /\/help\/gw$/],
                ['About', /\/about$/],
                ['Wiring', /\/w$/],
            ]
            await page.goto('/about')
            for (const [label, url] of views) {
                await openDrawer(page)
                await page.getByRole('link', { name: label, exact: true }).click()
                await expect(page).toHaveURL(url)
                await expect(main(page)).not.toBeEmpty()
            }
        })

        test('navigates to a containee in the wiring view', async ({ page }) => {
            const id = netnsIdOf('mock_bar_1')
            await gotoAndDiscover(page, '/w')
            await openDrawer(page)
            await page.getByRole('link').filter({ hasText: 'mock_bar_1' }).click()
            await expect(page).toHaveURL(new RegExp(`/w#.*netns-${id}$`))
            await expect(netnsCard(page, id)).toBeInViewport()
        })

        test('filters containees', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            await openDrawer(page)
            await page.getByPlaceholder('filter').fill('nostalgic')
            await page.keyboard.press('Escape')
            await expect(main(page)).toContainText('filtering applied')
            await expect(netnsCard(page, netnsIdOf('nostalgic_shannon'))).toBeVisible()
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toHaveCount(0)
        })

        test('filters by composer project name', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            await openDrawer(page)
            await page.getByPlaceholder('filter').fill('mock')
            await page.keyboard.press('Escape')
            await expect(main(page)).toContainText('filtering applied')
            // "edge-iot-core" doesn't contain its composer project name
            // "mock", so its IP stack can only match via the project name.
            await expect(netnsCard(page, netnsIdOf('edge-iot-core'))).toBeVisible()
            await expect(netnsCard(page, netnsIdOf('nostalgic_shannon'))).toHaveCount(0)
        })

        test('shows only the matching members of a composer project', async ({ page }) => {
            await gotoAndDiscover(page, '/w')
            await openDrawer(page)
            await page.getByPlaceholder('filter').fill('edge-iot')
            await page.keyboard.press('Escape')
            await expect(netnsCard(page, netnsIdOf('edge-iot-core'))).toBeVisible()
            // ...other members of the same "mock" composer project must now be
            // filtered out.
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toHaveCount(0)
        })

        test('applies project-aware filtering also in the details view', async ({ page }) => {
            await gotoAndDiscover(page, '/n')
            await openDrawer(page)
            await page.getByPlaceholder('filter').fill('edge-iot')
            await page.keyboard.press('Escape')
            await expect(netnsCard(page, netnsIdOf('edge-iot-core'))).toBeAttached()
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toHaveCount(0)
            await openDrawer(page)
            await page.getByPlaceholder('filter').fill('mock')
            await page.keyboard.press('Escape')
            // mock_bar_1 only reappears after the changed filter has been
            // applied; edge-iot-core then proves project name matching.
            await expect(netnsCard(page, netnsIdOf('mock_bar_1'))).toBeAttached()
            await expect(netnsCard(page, netnsIdOf('edge-iot-core'))).toBeAttached()
        })

    })

})

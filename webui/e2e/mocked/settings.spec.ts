import { discover, gotoAndDiscover, main } from '../fixtures'
import { expect, hiddenNetnsIds, netnsCard, test } from './discovery'

test.describe('settings', () => {

    test('shows "lo"-only process network namespaces when enabled, persistently', async ({ page }) => {
        const id = hiddenNetnsIds[0]
        await page.goto('/settings')
        await main(page).getByRole('listitem')
            .filter({ hasText: 'Show non-container "isolated" network namespaces' })
            .locator('input[type="checkbox"]').check()
        await gotoAndDiscover(page, '/w')
        await expect(netnsCard(page, id)).toBeVisible()
        await page.reload()
        await discover(page)
        await expect(netnsCard(page, id)).toBeVisible()
    })

    test('switches the theme', async ({ page }) => {
        await page.emulateMedia({ colorScheme: 'light' })
        await page.goto('/settings')
        const body = page.locator('body')
        const lightBackground = await body.evaluate(el => getComputedStyle(el).backgroundColor)
        await main(page).getByRole('listitem').filter({ hasText: 'Theme' }).getByRole('combobox').click()
        await page.getByRole('option', { name: 'dark' }).click()
        await expect(body).not.toHaveCSS('background-color', lightBackground)
    })

})

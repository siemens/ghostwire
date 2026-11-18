import { gotoAndDiscover, main, openDrawer, tocTitles } from '../fixtures'
import { expect, test } from '../mocked/discovery'

// Tests the production build with the e2e/branded/overlay/config.tsx app
// configuration, see playwright.config.ts.

const brand = 'Example Netview'

test.describe('non-default theme and branding', () => {

    test('brands the app bar, title, and favicon', async ({ page }) => {
        await gotoAndDiscover(page, '/w')
        await expect(page).toHaveTitle(brand)
        const banner = page.getByRole('banner')
        await expect(banner).toContainText(brand)
        await expect(banner.getByTestId('example-brand-icon')).toBeVisible()
        // small assets get inlined as data URLs.
        await expect(page.locator('link[rel~="icon"]')).toHaveAttribute('href', /6a1b9a/)
    })

    test('serves its own page shell', async ({ page, request }) => {
        const shell = await (await request.get('/')).text()
        expect(shell).toContain('<title>Example Netview</title>')
        expect(shell).toContain('<meta name="description" content="Example Netview virtual network discovery" />')
        expect(shell).toContain('You need to enable JavaScript to run the Example Netview app.')
        expect(shell).toMatch(/<script type="module" crossorigin src="\.\/assets\/index-[^"]+\.js">/)
        expect(shell).not.toContain('Ghostwire')
        await page.goto('/w')
        await expect(page.locator('#example-spinner')).toBeHidden()
        await expect(page.getByRole('banner')).toContainText(brand)
    })

    test('applies the light theme options', async ({ page }) => {
        await page.emulateMedia({ colorScheme: 'light' })
        await gotoAndDiscover(page, '/w')
        await expect(page.locator('header.MuiAppBar-root')).toHaveCSS('background-color', 'rgb(106, 27, 154)')
        await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(243, 229, 245)')
    })

    test('applies the dark theme options', async ({ page }) => {
        await page.emulateMedia({ colorScheme: 'dark' })
        await page.goto('/settings')
        const toggle = main(page).locator('.MuiSwitch-switchBase').first()
        await toggle.click()
        await expect(toggle).toHaveClass(/Mui-checked/)
        await expect(toggle).toHaveCSS('color', 'rgb(206, 147, 216)')
        // Ghostwire's dark background stays, as it isn't overridden.
        await expect(page.locator('body')).toHaveCSS('background-color', 'rgb(48, 48, 48)')
    })

    test('shows the about content, version, and legal links', async ({ page }) => {
        await gotoAndDiscover(page, '/about')
        await expect(main(page).getByRole('heading').first()).toHaveText(brand)
        await expect(main(page)).toContainText('An example downstream distribution')
        await expect(main(page)).toContainText('web app version: 9.8.7-example')
        const links = main(page).getByRole('list', { name: 'legal information' }).getByRole('link')
        await expect(links).toHaveText(['Imprint', 'Open Source Software Notices'])
        await expect(links.first()).toHaveAttribute('href', 'https://example.org/imprint')
    })

    test('brands the help and drawer', async ({ page }) => {
        await page.goto('/help/gw')
        expect((await tocTitles(page))[0]).toBe(brand)
        await openDrawer(page)
        await expect(page.getByRole('link', { name: 'About' }).getByTestId('example-brand-icon')).toBeVisible()
    })

})

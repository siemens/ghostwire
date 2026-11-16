import { test as base, expect, type Page } from '@playwright/test'

export { expect }

type Options = {
    /**
     * console error messages and uncaught exception messages expected by a
     * test that thus must not fail it.
     */
    allowedConsoleErrors: RegExp[]
}

type Fixtures = {
    failOnErrors: void
}

/**
 * Fails tests on uncaught exceptions and console errors, as otherwise crashes
 * might go unnoticed behind error boundary fallbacks; React logs the errors
 * caught by error boundaries to the console.
 */
export const test = base.extend<Options & Fixtures>({
    allowedConsoleErrors: [[], { option: true }],
    failOnErrors: [async ({ page, allowedConsoleErrors }, use) => {
        const errors: string[] = []
        const allowed = (message: string) => allowedConsoleErrors.some(re => re.test(message))
        page.on('pageerror', err => {
            if (!allowed(err.message)) {
                errors.push(`uncaught: ${err.message}`)
            }
        })
        page.on('console', msg => {
            if (msg.type() === 'error' && !allowed(msg.text())) {
                errors.push(`console: ${msg.text()}`)
            }
        })
        await use()
        expect(errors, 'uncaught errors and console errors').toEqual([])
    }, { auto: true }],
})

export const main = (page: Page) => page.getByRole('main')

/**
 * Refreshes the discovery via the app bar button and waits for the discovery
 * API response, as automatic refreshing is off by default.
 */
export const discover = async (page: Page) => {
    const response = page.waitForResponse(resp => new URL(resp.url()).pathname.endsWith('/json'))
    await page.getByRole('banner').getByRole('button', { name: 'refresh', exact: true }).click()
    await response
}

/** Navigates to the specified path and then refreshes the discovery. */
export const gotoAndDiscover = async (page: Page, path: string) => {
    await page.goto(path)
    await discover(page)
}

export const openDrawer = async (page: Page) => {
    await page.getByRole('button', { name: 'navigation drawer' }).click()
}

/** Opens the help table of contents and returns its menu. */
export const openToc = async (page: Page) => {
    await main(page).getByRole('button', { name: 'open table of contents' }).click()
    const menu = page.getByRole('menu')
    await expect(menu).toBeVisible()
    return menu
}

export const closeToc = async (page: Page) => {
    await page.keyboard.press('Escape')
    await expect(page.getByRole('menu')).toBeHidden()
}

/** Returns the chapter titles from the help table of contents, in order. */
export const tocTitles = async (page: Page) => {
    const menu = await openToc(page)
    const titles = await menu.getByRole('menuitem').allTextContents()
    await closeToc(page)
    return titles
}

/** Waits for the lazily loaded help chapter to have been fully rendered. */
export const expectChapterRendered = async (page: Page) => {
    await expect(main(page).getByRole('heading').first()).toHaveText(/\S/)
    await expect(main(page).locator('.MuiSkeleton-root')).toHaveCount(0)
}

/** Expects the table of contents to mark the specified chapter as current. */
export const expectCurrentChapter = async (page: Page, title: string) => {
    const menu = await openToc(page)
    await expect(menu.locator('.Mui-selected')).toHaveText(title)
    await closeToc(page)
}

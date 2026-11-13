import { readdirSync } from 'node:fs'
import path from 'node:path'

import type { Page } from '@playwright/test'

import {
    expectChapterRendered, expectCurrentChapter, main, openToc, tocTitles,
} from '../fixtures'
import { expect, test } from './discovery'

const chapterFiles = readdirSync(
    path.join(import.meta.dirname, '../../src/views/help/chapters'))
    .filter(name => name.endsWith('.mdx'))

const viewPaths = ['/w', '/n', '/lochla', '/settings', '/about']

/** Returns the same-origin link paths in the current chapter not in validPaths. */
const invalidChapterLinks = async (page: Page, validPaths: Set<string>) => {
    const origin = new URL(page.url()).origin
    const hrefs = await main(page).locator('a[href]').evaluateAll(
        anchors => anchors.map(a => (a as HTMLAnchorElement).href))
    return hrefs
        .map(href => new URL(href))
        .filter(url => url.origin === origin && !validPaths.has(url.pathname))
        .map(url => url.pathname)
}

test.describe('help', () => {

    test('has a table of contents entry for every chapter file', async ({ page }) => {
        await page.goto('/help')
        expect(await tocTitles(page)).toHaveLength(chapterFiles.length)
    })

    test('renders and deep-links all chapters, with valid in-chapter links', async ({ page }) => {
        test.slow()
        await page.goto('/help')
        const titles = await tocTitles(page)

        const chapterPaths: string[] = []
        for (const title of titles) {
            await test.step(`select "${title}" from table of contents`, async () => {
                const menu = await openToc(page)
                await menu.getByRole('menuitem', { name: title, exact: true }).click()
                await expect(page).toHaveURL(/\/help\/[^/]+$/)
                await expectChapterRendered(page)
                chapterPaths.push(new URL(page.url()).pathname)
                await expectCurrentChapter(page, title)
            })
        }
        expect(new Set(chapterPaths).size, 'unique chapter paths').toBe(titles.length)

        const validPaths = new Set([...viewPaths, ...chapterPaths])
        for (const [idx, chapterPath] of chapterPaths.entries()) {
            await test.step(`deep-link to ${chapterPath}`, async () => {
                await page.goto(chapterPath)
                await expectChapterRendered(page)
                await expectCurrentChapter(page, titles[idx])
                expect.soft(await invalidChapterLinks(page, validPaths),
                    `invalid links in chapter "${titles[idx]}"`).toEqual([])
            })
        }
    })

    test('walks through all chapters forward and backward', async ({ page }) => {
        test.slow()
        await page.goto('/help')
        const titles = await tocTitles(page)
        const prev = main(page).locator('button.prev')
        const next = main(page).locator('button.next')

        await expectChapterRendered(page)
        await expect(prev).toHaveCount(0)
        for (const title of titles.slice(1)) {
            await expect(next).toHaveText(title)
            await next.click()
            await expectChapterRendered(page)
        }
        await expect(next).toHaveCount(0)
        for (const title of titles.slice(0, -1).reverse()) {
            await expect(prev).toHaveText(title)
            await prev.click()
            await expectChapterRendered(page)
        }
        await expect(prev).toHaveCount(0)
    })

    test('falls back to the first chapter for unknown chapters', async ({ page }) => {
        await page.goto('/help')
        const [first] = await tocTitles(page)
        await page.goto('/help/doesnotexist')
        await expectChapterRendered(page)
        await expectCurrentChapter(page, first)
    })

    test('follows in-chapter links to views', async ({ page }) => {
        await page.goto('/help/details')
        await main(page).getByRole('link', { name: 'settings', exact: true }).first().click()
        await expect(page).toHaveURL(/\/settings$/)
        await expect(main(page)).toContainText('Appearance')
    })

})

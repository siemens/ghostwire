import {
    expect, expectChapterRendered, expectCurrentChapter, gotoAndDiscover, main, openDrawer, test,
    tocTitles,
} from '../fixtures'
import { brand, containees, forwardedPort } from './setup'

test.describe('live Ghostwire service', () => {

    test('applies the brand', async ({ page }) => {
        await page.goto('/')
        await expect(page).toHaveTitle(brand)
        await expect(page.getByRole('banner')).toContainText(brand)
        await page.goto('/help')
        expect((await tocTitles(page))[0]).toBe(brand)
        await expect(main(page).getByRole('heading').first()).toHaveText(brand)
    })

    test('serves deep links to help chapters', async ({ page }) => {
        await page.goto('/help/wiring')
        await expectChapterRendered(page)
        await expectCurrentChapter(page, 'Wiring View')
    })

    test('shows the well-known containees in the wiring view', async ({ page }) => {
        await gotoAndDiscover(page, '/w')
        for (const name of containees) {
            await expect(main(page).getByRole('button').filter({ hasText: name })).toBeVisible()
        }
    })

    test('serves deep links to containee details', async ({ page }) => {
        await gotoAndDiscover(page, `/n/${containees[0]}`)
        await expect(main(page).getByRole('link', { name: 'back to overall view' })).toBeVisible()
        await expect(main(page)).toContainText(containees[0])
        await expect(main(page)).toContainText('eth0')
    })

    test('lists the forwarded host port', async ({ page, request }) => {
        type JsonNetns = {
            containers: { pid: number }[]
            'forwarded-ports': Record<string, { port: number }[]>
        }
        const netnses = (await (await request.get('/json')).json())['network-namespaces'] as JsonNetns[]
        const fwdNetns = netnses.find(netns => Object.values(netns['forwarded-ports'])
            .flat().some(fwd => fwd.port === forwardedPort))
        expect(fwdNetns, 'network namespace forwarding the port').toBeDefined()
        test.skip(!fwdNetns!.containers.some(cntr => cntr.pid === 1),
            'container engine not in initial network namespace, such as in a dev container')

        await gotoAndDiscover(page, '/lochla')
        await expect(main(page).getByRole('row')
            .filter({ has: page.getByRole('cell', { name: `:${forwardedPort}`, exact: true }) }))
            .toContainText(':8080')
    })

    test('shows the discovery engine version', async ({ page }) => {
        await gotoAndDiscover(page, '/about')
        await expect(main(page)).toContainText(/discovery backend engine: gostwire \S+/)
    })

    test.describe('behind a path-rewriting proxy', () => {

        const prefix = '/edgeshark'

        // Emulates an ingress serving only below the prefix, which it strips
        // before forwarding with an X-Forwarded-Prefix header.
        test.beforeEach(async ({ page, baseURL }) => {
            const origin = new URL(baseURL!).origin
            await page.route('**/*', async route => {
                const url = new URL(route.request().url())
                if (url.origin !== origin) {
                    return route.continue()
                }
                if (!url.pathname.startsWith(`${prefix}/`)) {
                    return route.fulfill({ status: 404, body: 'ingress: not found' })
                }
                url.pathname = url.pathname.slice(prefix.length)
                return route.fulfill({
                    response: await route.fetch({
                        url: url.toString(),
                        headers: { ...route.request().headers(), 'x-forwarded-prefix': prefix },
                    }),
                })
            })
        })

        test('discovers via the prefixed API', async ({ page }) => {
            await gotoAndDiscover(page, `${prefix}/w`)
            await expect(main(page).getByRole('button').filter({ hasText: containees[0] })).toBeVisible()
        })

        test('keeps help chapter links below the prefix', async ({ page }) => {
            await page.goto(`${prefix}/help/gw`)
            await expectChapterRendered(page)
            const origin = new URL(page.url()).origin
            const paths = (await main(page).locator('a[href]').evaluateAll(
                anchors => anchors.map(a => (a as HTMLAnchorElement).href)))
                .map(href => new URL(href))
                .filter(url => url.origin === origin)
                .map(url => url.pathname)
            expect(paths).not.toHaveLength(0)
            for (const path of paths) {
                expect(path).toMatch(new RegExp(`^${prefix}/`))
            }
            await main(page).getByRole('link', { name: 'Technical Features' }).click()
            await expect(page).toHaveURL(new RegExp(`${prefix}/help/tech$`))
            await expectCurrentChapter(page, 'Technical Features')
        })

        test('links from about to help below the prefix', async ({ page }) => {
            await page.goto(`${prefix}/about`)
            await main(page).getByRole('link', { name: 'help', exact: true }).click()
            await expect(page).toHaveURL(new RegExp(`${prefix}/help/gw$`))
            await expectChapterRendered(page)
        })

        test('navigates views below the prefix', async ({ page }) => {
            await page.goto(`${prefix}/w`)
            await openDrawer(page)
            await page.getByRole('link', { name: 'Settings', exact: true }).click()
            await expect(page).toHaveURL(new RegExp(`${prefix}/settings$`))
            await expect(main(page)).toContainText('Appearance')
        })

    })

})

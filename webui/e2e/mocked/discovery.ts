import { readFileSync } from 'node:fs'
import path from 'node:path'

import type { Page } from '@playwright/test'

import { test as base } from '../fixtures'

export { expect } from '../fixtures'

type JsonNetns = {
    netnsid: number
    containers: { name: string, type: string }[]
    'network-interfaces': { name: string }[]
}

/** The discovery data served instead of a real Ghostwire service. */
export const discovery = JSON.parse(readFileSync(
    path.join(import.meta.dirname, '../../src/models/gw/mock/mockdata.json'), 'utf8'))

const netnses = discovery['network-namespaces'] as JsonNetns[]

const isEmptyNonContainerNetns = (netns: JsonNetns) =>
    netns['network-interfaces'].length === 1
    && netns['network-interfaces'][0].name === 'lo'
    && netns.containers.every(c => c.type === 'proc')

/** IDs of the network namespaces shown by default. */
export const shownNetnsIds = netnses.filter(n => !isEmptyNonContainerNetns(n)).map(n => n.netnsid)

/** IDs of the "lo"-only process network namespaces hidden by default. */
export const hiddenNetnsIds = netnses.filter(isEmptyNonContainerNetns).map(n => n.netnsid)

export const netnsIdOf = (containeeName: string) => {
    const netns = netnses.find(n => n.containers.some(c => c.name === containeeName))
    if (!netns) {
        throw new Error(`containee "${containeeName}" not in discovery fixture`)
    }
    return netns.netnsid
}

/** Locates the anchor of a network namespace card, regardless of ID prefixes. */
export const netnsCard = (page: Page, netnsid: number) =>
    page.locator(`[id$="netns-${netnsid}"]`)

type Options = {
    /** HTTP status of the discovery API responses; 200 serves the fixture. */
    discoveryStatus: number
}

export const test = base.extend<Options>({
    discoveryStatus: [200, { option: true }],
    page: async ({ page, discoveryStatus }, use) => {
        await page.route(url => url.pathname === '/json', route => route.fulfill(
            discoveryStatus === 200 ? { json: discovery } : { status: discoveryStatus }))
        await use(page)
    },
})

// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { getFilterFn } from '../filterinput'
import { AddressFamily, emptyNonContainerNetns, netnsMatchesFilter, orderNetnsByContainees } from '../../models/gw'
import { discovery, loOnlyContainerNetns } from '../../models/gw/mock'

import { NetnsBreadboard } from './NetnsBreadboard'

// Only some network namespaces, so that some wires lead to network interfaces
// that aren't rendered at all.
const somenetnses = Object.values(discovery.networkNamespaces)
    .sort(orderNetnsByContainees)
    .filter(netns => Object.values(netns.nifs).length === 1
        || ['init', 'ghostwire'].some(prefix => netns.containers[0].name.startsWith(prefix)))
    .concat(loOnlyContainerNetns)

// The network namespaces matching a containee filter pattern, narrowed down
// the same way as the wiring view does it: callers apply the containee filter
// and the breadboard then renders exactly the matching network namespaces,
// with composer project groups covering only the matching members.
const matchingNetnses = (pattern: string) => {
    const filterfn = getFilterFn({ pattern, isCaseSensitive: false, isRegexp: false })
    return Object.values(discovery.networkNamespaces)
        .filter(netns => !emptyNonContainerNetns(netns))
        .filter(netns => netnsMatchesFilter(netns, filterfn))
        .sort(orderNetnsByContainees)
}

const meta: Meta<typeof NetnsBreadboard> = {
    title: 'Ghostwire/NetnsBreadboard',
    component: NetnsBreadboard,
    args: {
        netns: {
            ...discovery.networkNamespaces,
            [loOnlyContainerNetns.netnsid]: loOnlyContainerNetns,
        },
        // mirror the app's default of hiding lo.
        filterLo: true,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NetnsBreadboard>

export const Everything: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const WithLo: Story = {
    args: {
        filterLo: false,
    },
}

/**
 * This story has filtering enabled:
 *
 * - filter out (hide) `lo` loopback network interfaces.
 * - filter out (hide) non-container network namespaces with only a single,
 *   lonely loopback network interface; the "network:none" container stays
 *   visible including its lo.
 *
 * Please note that this story filters out containers in order to reduce the
 * number of network namespaces rendered, except for a few white-listed
 * containers only and namespaces with only a loopback network interface. The
 * latter namespaces are then filtered out in the final rendering using the
 * `filterEmpty` property of the `NetnsBreadboard` component.
 *
 * A nice side effect of this story is that it shows that the wire rendering
 * should correctly work with incomplete wiring, where some network
 * peer/master/... network interfaces in a related set of network interfaces
 * haven't been rendered.
 */
export const Filtered: Story = {
    args: {
        netns: somenetnses,
        filterEmpty: true,
    },
}

/**
 * Containee filtering happens in the caller, as in the wiring view: the
 * pattern "edge-iot" matches only the "edge-iot-core" container, so its
 * composer project card shows just this single matching member instead of the
 * complete "mock" project.
 */
export const FilteredByContaineeName: Story = {
    args: {
        netns: matchingNetnses('edge-iot'),
    },
}

/**
 * The containee filter also matches composer project names: the pattern "mock"
 * reveals the complete "mock" composer project, including the "edge-iot-core"
 * member whose container name doesn't contain the pattern at all.
 */
export const FilteredByProjectName: Story = {
    args: {
        netns: matchingNetnses('mock'),
    },
}

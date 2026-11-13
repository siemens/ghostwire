// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { ReactNode } from 'react'

import type { Decorator, Meta, StoryObj } from '@storybook/react-vite'

import { Provider, type WritableAtom } from 'jotai'
import { useHydrateAtoms } from 'jotai/utils'

import { discovery } from 'models/gw/mock'
import { filterCaseSensitiveAtom, filterPatternAtom, filterRegexpAtom, persistSettingsAtom } from 'views/settings'

import { ContaineeNavigator } from './ContaineeNavigator'

interface HydrateAtomsProps {
    initialValues: [WritableAtom<unknown, any[], any>, unknown][] // eslint-disable-line @typescript-eslint/no-explicit-any
    children: ReactNode
}

const HydrateAtoms = ({ initialValues, children }: HydrateAtomsProps) => {
    useHydrateAtoms(initialValues)
    return children
}

// Scopes the (user settings) filter atoms to the individual story, supplying a
// deterministic, non-persisted containee filter pattern.
const withFilterPattern = (pattern: string): Decorator => {
    const FilterPatternDecorator: Decorator = (Story) => (
        <Provider>
            <HydrateAtoms initialValues={[
                [persistSettingsAtom, false],
                [filterPatternAtom, pattern],
                [filterCaseSensitiveAtom, false],
                [filterRegexpAtom, false],
            ]}>
                <Story />
            </HydrateAtoms>
        </Provider>
    )
    return FilterPatternDecorator
}

const meta: Meta<typeof ContaineeNavigator> = {
    title: 'Ghostwire/ContaineeNavigator',
    component: ContaineeNavigator,
    args: {
        allnetns: discovery.networkNamespaces,
    },
    argTypes: {
        allnetns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContaineeNavigator>

export const Standard: Story = {}

export const FilterEmpty: Story = {
    args: {
        filterEmpty: true,
    },
}

export const NoLinks: Story = {
    args: {
        nolink: true,
    },
}

/**
 * The containee filter is applied per network namespace, using the same
 * matching rules as in the wiring and details views (`netnsMatchesFilter`):
 * the pattern "edge-iot" here matches the composed "edge-iot-core" container
 * by its container name, so only the containees of its network namespace get
 * listed.
 */
export const FilterByContaineeName: Story = {
    decorators: [withFilterPattern('edge-iot')],
}

/**
 * The containee filter also matches composer project names: the
 * "edge-iot-core" container gets listed even though its name doesn't contain
 * the pattern "mock", because it is a member of the "mock" composer project.
 */
export const FilterByProjectName: Story = {
    decorators: [withFilterPattern('mock')],
}

// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { DynVarsProvider, useDynVars } from './DynVars'

const FooBar = ({ text }: { text: string }) => {
    const dynvars = useDynVars()
    return <div>{text}: {dynvars.foobar ? String(dynvars.foobar) : <em>nothing</em>}</div>
}

const meta: Meta<typeof DynVarsProvider> = {
    title: 'Universal/DynVarsProvider',
    component: DynVarsProvider,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof DynVarsProvider>

export const Override: Story = {
    render: () => <>
        <FooBar text="default dynvars" />
        <DynVarsProvider value={{ foobar: 'hooray!' }}>
            <FooBar text="locally overridden dynvars" />
        </DynVarsProvider>
    </>,
}

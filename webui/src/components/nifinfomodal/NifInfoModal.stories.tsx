// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Button, Stack } from '@mui/material'

import { nifBr, nifHw, nifLo, nifVeth1 } from 'models/gw/mock'

import { useNifInfoModal } from './context'
import { NifInfoModalProvider } from './NifInfoModal'

const ShowInfo = () => {
    const setNif = useNifInfoModal()
    return <Stack direction="row" spacing={1}>
        {[nifHw, nifVeth1, nifBr, nifLo].map(nif =>
            <Button key={nif.name} variant="outlined" onClick={() => setNif?.(nif)}>
                {nif.name}
            </Button>)}
    </Stack>
}

const meta: Meta<typeof NifInfoModalProvider> = {
    title: 'Ghostwire/NifInfoModalProvider',
    component: NifInfoModalProvider,
    args: {
        children: <ShowInfo />,
    },
    argTypes: {
        children: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifInfoModalProvider>

/** Descendants open the modal using the setter returned by `useNifInfoModal()`. */
export const Standard: Story = {}

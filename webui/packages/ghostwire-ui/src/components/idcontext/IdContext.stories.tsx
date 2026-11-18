// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box, Card, Stack } from '@mui/material'

import { useContextualId } from './useidcontext'
import { IdContext } from './IdContext'

const ShowIdContext = ({ id }: { id: string }) => {
    const uid = useContextualId(id)
    return <div>DOM element identifier is "{uid}".</div>
}

const meta: Meta<typeof IdContext> = {
    title: 'Universal/IdContext',
    component: IdContext,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof IdContext>

/**
 * This story demonstrates how to introduce separate DOM element identifier
 * contexts, or prefixes. The decorator wrapping these stories automatically
 * establishes identifier contexts for each story individually. However, in
 * your own applications you need to establish different contexts "manually" as
 * required.
 */
export const SeparateContexts: Story = {
    render: () => (
        <Stack spacing={2} sx={{ p: 2 }}>
            <Card>
                <Box sx={{ p: 1 }}>
                    <IdContext>
                        <ShowIdContext id="A" />
                        <ShowIdContext id="B" />
                    </IdContext>
                </Box>
            </Card>
            <Card>
                <Box sx={{ p: 1 }}>
                    <IdContext>
                        <ShowIdContext id="A" />
                    </IdContext>
                </Box>
            </Card>
        </Stack>
    ),
}

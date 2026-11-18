// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box, Card } from '@mui/material'

import { CardSection } from './CardSection'

const meta: Meta<typeof CardSection> = {
    title: 'Universal/CardSection',
    component: CardSection,
    decorators: [
        (Story) => <Card>
            <Box sx={{ p: 2, width: '20em' }}>
                <Story />
            </Box>
        </Card>,
    ],
    args: {
        caption: 'dixit',
        children: 'lorem ipsum dolor sit amet, consetetur sadipsicing eliter, '
            + 'sed superbus esse semper latinum labere.',
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof CardSection>

/**
 * Unless specified otherwise, a card section shows its caption as well as its
 * children.
 */
export const Standard: Story = {}

/**
 * However, card sections can be made collapsible/expandable by specifying
 * their initial `collapsible=` state `expand` or `collapse` (but not `never`).
 * Users then can collapse or expand a section as to their hearts' desire.
 */
export const Expanded: Story = {
    args: {
        collapsible: 'expand',
    },
}

export const Collapsed: Story = {
    args: {
        collapsible: 'collapse',
    },
}

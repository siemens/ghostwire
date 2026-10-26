// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { styled } from '@mui/material'

/** Reflowing grid of equal-width `Tile`s. */
export const Tiles = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fill, minmax(11em, 1fr))',
    gap: theme.spacing(2),
    // blend the gaps with the canvas surrounding the (paper) story area.
    background: theme.palette.background.default,
}))

/** Outlined tile, typically a `Typography` caption followed by the showcased item. */
export const Tile = styled('div')(({ theme }) => ({
    padding: theme.spacing(1),
    border: `1px solid ${theme.palette.divider}`,
    borderRadius: theme.shape.borderRadius,
    background: theme.palette.background.paper,
    '& > .MuiTypography-root': {
        marginBottom: '0.5lh',
    },
}))

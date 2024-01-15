// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Skeleton, styled, Theme, Typography } from '@mui/material'
import { SxProps } from '@mui/system'

const Bones = styled('div')(() => ({
    width: '100%',
    '& > h4:nth-of-type(1)': { width: '55%' },
    '& > p:nth-of-type(1)': { width: '90%' },
    '& > p:nth-of-type(2)': { width: '100%' },
    '& > p:nth-of-type(3)': { width: '70%' },
}))


export interface ChapterSkeletonProps {
    /**
     * The MUI system prop that allows defining system overrides as well as
     * additional CSS styles.
     */
    sx?: SxProps<Theme>
}

/**
 * Renders a simple chapter-like text skeleton. It can be used as a fallback
 * display while MDX modules are getting lazily loaded, so users don't see just
 * a blank screen, but instead get some visual feedback of a pending operation.
 */
export const ChapterSkeleton = ({sx}: ChapterSkeletonProps) => {
    return <Bones sx={sx}>
        <Typography variant="h4"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
        <Typography variant="body1"><Skeleton animation="wave" /></Typography>
    </Bones>
}

export default ChapterSkeleton

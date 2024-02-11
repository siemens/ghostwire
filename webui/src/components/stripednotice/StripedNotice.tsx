// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { ReactNode } from 'react'
import { styled } from '@mui/material'
import { rgba } from 'utils/rgba'

const Container = styled('div')(({ theme }) => ({
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    paddingLeft: theme.spacing(1),
    paddingRight: theme.spacing(1),
}))

const Message = styled('div')(({ theme }) => ({
    color: rgba(theme.palette.text.secondary, 0.5),
    paddingLeft: theme.spacing(2),
    paddingRight: theme.spacing(2),
}))

const Stripe = styled('div')(({ theme }) => ({
    flex: 1,
    height: '1rem',
    background: `repeating-linear-gradient(`
        + `-55deg, `
        + `${rgba(theme.palette.text.secondary, 0.3)}, `
        + `${rgba(theme.palette.text.secondary, 0.3)} 10px, `
        + `${theme.palette.background.default} 10px, `
        + `${theme.palette.background.default} 20px)`,
}))

export interface StripedNoticeProps {
    /** mandatory children elements */
    children: ReactNode
}

/**
 * `StripedNotice` renders a horizontal "hazzard stripe"-type bar, with a
 * centered message.
 */
export const StripedNotice = ({ children }: StripedNoticeProps) => {
    return <Container>
        <Stripe />
        <Message>{children}</Message>
        <Stripe />
    </Container>
}

export default StripedNotice

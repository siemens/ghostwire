// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { Link } from 'react-router-dom'

import { ExtLink } from 'components/extlink'


export interface SmartAProps {
    /** hyper reference */
    href: string
    /** children to render inside the hyperlink. */
    children: React.ReactNode
}
const isAbsoluteURL = (href: string) => {
    try {
        new URL(href)
        return true
    } catch {
        return false
    }
}

/**
 * Renders a hyperlink either as an external link (using the ExtLink component),
 * or a React (DOM) router "internal" Link component, depending on the given
 * href property value. Using the Link component ensures proper app-internal
 * route handling without having to reload the application and thus destroying
 * the any discovery result.
 */
export const SmartA = ({href, children, ...otherprops}: SmartAProps) => (
    isAbsoluteURL(href)
        ? <ExtLink href={href} {...otherprops}>{children}</ExtLink>
        : <Link to={href} {...otherprops}>{children}</Link>
)

export default SmartA

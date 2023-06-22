// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { ExtLink } from 'components/extlink'
import React from 'react'
import { Link } from 'react-router-dom'


export interface SmartAProps {
    /** hyper reference */
    href: string
    /** children to render inside the hyperlink. */
    children: React.ReactNode
}

/**
 * Renders a hyperlink either as an external link (using the ExtLink component),
 * or a react router "internal" Link component, depending on the given href
 * property value. Using the Link component ensures proper app-internal route
 * handling without having to reload the application and thus destroying the any
 * discovery result.
 */
export const SmartA = ({href, children, ...otherprops}: SmartAProps) => {
    try {
        new URL(href)
        return <ExtLink href={href} {...otherprops}>{children}</ExtLink>
    } catch {
        return <Link to={href} {...otherprops}>{children}</Link>
    }
}

export default SmartA

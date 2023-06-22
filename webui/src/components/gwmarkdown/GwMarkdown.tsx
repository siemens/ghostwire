// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { MuiMarkdown } from 'components/muimarkdown'
import { Link } from 'react-router-dom'
import { ExtLink } from 'components/extlink'
import { styled } from '@mui/material';


const GwMD = styled(MuiMarkdown)(({ theme }) => ({
    // while we allow badges to wrap in the views, we don't want that to
    // happen in our exemplary tables, so with enforce no wrapping in the
    // about documentation.
    '& div[class*="-badge-"]': {
        whiteSpace: 'nowrap',
    },

    '& .MuiTypography-h1, & .MuiTypography-h2, & .MuiTypography-h3, & .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6, & .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
        color: theme.palette.mode === 'light' ? theme.palette.primary.dark : theme.palette.primary.light,
    },
    '& .MuiTypography-body2': {
        borderLeftColor: theme.palette.mode === 'light' ? theme.palette.primary.dark : theme.palette.primary.light,
    },
    '& a, & a:visited': {
        color: theme.palette.mode === 'light' ? theme.palette.primary.dark : theme.palette.primary.light,
    },
    '& a:active': {
        color: theme.palette.secondary.main,
    },
}))


/**
 * Renders a hyperlink either as an external link (using the ExtLink component),
 * or a react router "internal" Link component, depending on the given href
 * property value. Using the Link component ensures proper app-internal route
 * handling without having to reload the application and thus destroying the any
 * discovery result.
 */
const SmartA = ({ href, children, ...otherprops }: any) => {
    try {
        new URL(href)
        return <ExtLink href={href} {...otherprops}>{children}</ExtLink>
    } catch {
        return <Link to={href} {...otherprops}>{children}</Link>
    }
}


export interface GwMarkdownProps {
    /** compiled MDX, which can also be lazy loaded. */
    mdx: (props: any) => JSX.Element
    /** 
     * an object "map" of "shortcodes" (which is a rather fancy name for
     * "components") to be made available to the MDX without the need to
     * explicitly import them in the MDX.
     */
    shortcodes?: { [key: string]: React.ComponentType<any> }
    /** CSS class name(s). */
    className?: string
    /** fallback components to render when lazily loading the mdx. */
    fallback?: JSX.Element
}

/**
 * Renders the given MDX using Material-UI typography components and applies
 * additional Ghostwire-specific styling.
 *
 * For convenience, this `GwMarkdown` component renders hyperlinks differently,
 * depending on whether they are relative or absolute hyperlinks:
 * - relative hyperlinks are considered app-internal and are thus rendering
 *   using the `Link` component from react-router.
 * - absolute hyperlinks are rendering using the `ExtLink` component; this will
 *   render an "external link" icon as well as always open a new tab/window and
 *   will have noopener and noreferrer enforced.
 */
export const GwMarkdown = ({ mdx, className, shortcodes, fallback }: GwMarkdownProps) => {
    return <GwMD
        className={className}
        shortcodes={{ a: SmartA, ...shortcodes }}
        mdx={mdx}
        fallback={fallback}
    />
}

export default GwMarkdown

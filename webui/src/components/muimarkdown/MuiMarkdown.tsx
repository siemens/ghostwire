// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { memo } from 'react'

import {
    Divider,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
    lighten,
    styled,
} from '@mui/material'
import { ChapterSkeleton } from 'components/chapterskeleton'
import { MDXContent } from 'mdx/types'


// Defines how to map the components emitted by MDX onto Material-UI components,
// and especially the Typography component. See also:
// https://mdxjs.com/advanced/components
const muiComponents = {
    // Get us rid of that pesky "validateDOMNesting(...): <p> cannot appear as a
    // descendant of <p>" by using a <div> instead of Typography's default <p>.
    p: (() => {
        const P = (props: any) => <Typography {...props} component="div" />
        return memo(P)
    })(),

    h1: (() => {
        const H1 = (props: any) => <Typography {...props} variant="h4" />
        return memo(H1)
    })(),

    h2: (() => {
        const H2 = (props: any) => <Typography {...props} variant="h5" />
        return memo(H2)
    })(),

    h3: (() => {
        const H3 = (props: any) => <Typography {...props} variant="h6" />
        return memo(H3)
    })(),

    h4: (() => {
        const H4 = (props: any) => <Typography {...props} variant="subtitle1" />
        return memo(H4)
    })(),

    h5: (() => {
        const H5 = (props: any) => <Typography {...props} variant="subtitle2" />
        return memo(H5)
    })(),

    h6: (() => {
        const H6 = (props: any) => <Typography {...props} variant="subtitle2" />
        return memo(H6)
    })(),

    // And once more: get us rid of that pesky "validateDOMNesting(...): <p>
    // cannot appear as a descendant of <p>" by using a <div> instead of
    // Typography's default <p>.
    blockquote: (() => {
        const Blockquote = (props: any) => <Typography {...props} component="div" variant="body2" />
        return memo(Blockquote)
    })(),

    ul: (() => {
        const Ul = (props: any) => <Typography {...props} component="ul" />
        return memo(Ul)
    })(),

    ol: (() => {
        const Ol = (props: any) => <Typography {...props} component="ol" />
        return memo(Ol)
    })(),

    li: (() => {
        const Li = (props: any) => <Typography {...props} component="li" />
        return memo(Li)
    })(),

    table: (() => {
        const MuiTable = (props: any) => <Table {...props} />
        return memo(MuiTable)
    })(),

    tr: (() => {
        const Tr = (props: any) => <TableRow {...props} />
        return memo(Tr)
    })(),

    td: (() => {
        const Td = ({ align, ...props }) => (
            <TableCell align={align || undefined} {...props} />
        )
        return memo(Td)
    })(),

    tbody: (() => {
        const TBody = (props: any) => <TableBody {...props} />
        return memo(TBody)
    })(),

    th: (() => {
        const Th = ({ align, ...props }) => (
            <TableCell align={align || undefined} {...props} />
        )
        return memo(Th)
    })(),

    thead: (() => {
        const THead = (props: any) => <TableHead {...props} />
        return memo(THead)
    })(),

    hr: Divider,
}


// Styles Material-UIs typography elements inside am MDX context to our hearts'
// desires. Additionally styles some Mui components, such as Mui SVG icons to
// fit into the overall styling.
const MarkdownArea = styled('div')(({ theme }) => ({
    // Make sure to properly reset the text color according to the primary
    // text color.
    color: theme.palette.text.primary,
    // ...and now for the details...
    '& .MuiTypography-h1, & .MuiTypography-h2, & .MuiTypography-h3, & .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6, & .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
        color: theme.palette.mode === 'light'
            ? theme.palette.primary.main
            : theme.palette.primary.light,
    },
    '& .MuiTypography-h4:first-of-type': {
        marginTop: theme.spacing(1),
    },
    '& .MuiTypography-h4, & .MuiTypography-h5, & .MuiTypography-h6': {
        marginTop: theme.spacing(3),
        marginBottom: theme.spacing(2),
    },
    '& .MuiTypography-subtitle1, & .MuiTypography-subtitle2': {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(1),
    },
    '& .MuiTypography-body1 + .MuiTypography-body1': {
        marginTop: theme.spacing(1),
    },
    '& .MuiTypography-body2': {
        margin: theme.spacing(2),
        borderLeft: `${theme.spacing(1)} solid ${theme.palette.primary.main}`,
        paddingLeft: theme.spacing(1),
    },
    '& .MuiSvgIcon-root.icon': {
        verticalAlign: 'middle',
        fontSize: 'calc(100% + 2px)',
        border: `1px solid ${theme.palette.text.disabled}`,
        padding: 1,
        borderRadius: theme.spacing(0.5),
    },
    '& a:link': {
        color: theme.palette.mode === 'light'
            ? theme.palette.primary.main
            : theme.palette.primary.light
    },
    '& a:visited': {
        color: theme.palette.mode === 'light'
            ? theme.palette.primary.dark
            : lighten(theme.palette.primary.light, 0.3)
    },
    '& a:hover, & a:active': {
        color: theme.palette.secondary.main
    },
    '& code': {
        fontFamily: 'Roboto Mono',
    },
}))


export interface MuiMarkdownProps {
    /** compiled MDX, which can also be lazy loaded. */
    mdx: MDXContent
    /** shortcodes, that is, available components. */
    shortcodes?: { [key: string]: React.ComponentType<any> }
    /** CSS class name(s). */
    className?: string
    /** fallback components to render when lazily loading the mdx. */
    fallback?: JSX.Element
}

/**
 * Renders the given [MDX](https://mdxjs.com/) using Material-UI `Typography`
 * components (where appropriate). The MDX can be either statically imported
 * beforehand or alternatively lazily imported when needed using `React.lazy()`.
 * This component will handle both use cases transparently: it uses a
 * `React.Suspense` child component and shows a `ChapterSkeleton` component
 * while lazily loading MDX.
 *
 * - uses [mdx.js](https://github.com/mdx-js/mdx).
 * - headings automatically get `id` slugs via
 *   [remark-slug](https://github.com/remarkjs/remark-slug).
 * - some typography goodies via
 *   [remark-textr](https://github.com/remarkjs/remark-textr):
 *   - typographic ellipsis,
 *   - typgraphic quotes,
 *   - number range endashes,
 *   - turns `--` into emdashes.
 *
 * Please see the [`HelpViewer`](#helpviewer) component for a no-frills help
 * document viewer with multiple chapter support and chapter navigation.
 */
export const MuiMarkdown = ({ mdx: Mdx, className, shortcodes, fallback }: MuiMarkdownProps) => (
    <React.Suspense fallback={fallback || <ChapterSkeleton />}>
        <MarkdownArea className={className}>
            <Mdx components={{ ...muiComponents, ...shortcodes }} />
        </MarkdownArea>
    </React.Suspense>
)

export default MuiMarkdown

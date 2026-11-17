/**
 * Types MDX file imports (compiled by @mdx-js/rollup) as React components that
 * render the MDX content; props are passed on to the MDX content, such as
 * `components` for overriding the elements used for rendering.
 */
declare module '*.mdx' {
    let MDXComponent: (props: Record<string, unknown>) => import('react').JSX.Element
    export default MDXComponent
}

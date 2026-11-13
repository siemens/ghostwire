// Type declarations for the untyped textr plugins used by the MDX typography
// remark plugin in mdxconfig.ts; each plugin transforms a text string.

declare module 'typographic-apostrophes' {
    const plugin: (input: string) => string
    export default plugin
}

declare module 'typographic-quotes' {
    const plugin: (input: string, options?: { locale?: string }) => string
    export default plugin
}

declare module 'typographic-apostrophes-for-possessive-plurals' {
    const plugin: (input: string) => string
    export default plugin
}

declare module 'typographic-ellipses' {
    const plugin: (input: string) => string
    export default plugin
}

declare module 'typographic-en-dashes' {
    const plugin: (input: string) => string
    export default plugin
}

// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

// Consumers type-checking the library sources need the ambient MDX module
// declarations, which cannot be imported.
/// <reference types="vite/client" />
// eslint-disable-next-line @typescript-eslint/triple-slash-reference
/// <reference path="./mdx.d.ts" />

// The public API of the Ghostwire UI library.

export { default as App } from './app'
export * from './app'
export * from './components/brand'
export * from './components/brandicon'
export * from './components/branding'
export * from './components/dynvars'
export * from './styles/themes'

// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

/**
 * Determines this application's basename (path!) from the base path passed in
 * by the server in the dynamic variables, falling back to its <base href="...">
 * DOM element. The basename is "" if unset or "/", otherwise it's the specified
 * basename, but always without any trailing slash. This basename can thus be
 * directly fed into the basename property of a React DOM router component. For
 * this reason, the basename is stripped off of any scheme, host, port, hash,
 * and query elements.
 */
const serverBase = window.dynvars?.base

export const basename =
    new URL(
        (typeof serverBase === 'string' && serverBase)
        || (document.querySelector('base') || {}).href
        || '/',
        window.location.href,
    ).pathname
        // ensure that there is never a trailing slash, and this includes the root
        // itself.
        .replace(/\/$/, '')

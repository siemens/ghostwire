// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

/**
 * Determines this application's basename (path!) from its <base href="..."> DOM
 * element. The basename is "" if unset or "/", otherwise it's the specified
 * basename, but always without any trailing slash. This basename can thus be
 * directly fed into the basename property of a React DOM router component. For
 * this reason, the basename is stripped off of any scheme, host, port, hash,
 * and query elements.
 */
export const basename =
    new URL(
        // get the href attribute of the first base DOM element, falling back to "/"
        // if there isn't one.
        ((document.querySelector('base') || {}).href || '/')
    ).pathname
        // ensure that there is never a trailing slash, and this includes the root
        // itself.
        .replace(/\/$/, '')

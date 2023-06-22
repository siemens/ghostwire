// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'

/**
 * DynamicVars represents an object "map" of dynamic variable passed into an
 * application by the server at load time (as opposed to static REACT_APP_
 * variables which are fixed at build time).
 */
type DynamicVars = { [key: string]: any }

// Dynamic variables from the server get passed in via dynamically served
// "index.html" that sets the dynvars element of the window object.
declare global {
    interface Window {
        dynvars: DynamicVars
    }
}

/**
 * DynVarsContext is the context for accessing the dynamic variables passed in
 * to this application (if any). This context automatically defaults to the
 * dynamic variables passed in, so applications do not necessarily put in any
 * `<DynVarsProvider>` into their component tree, unless they want to override
 * any global dynamic variables, such as in the styleguide examples.
 */
const DynVarsContext = React.createContext(
    window.dynvars || ({} as DynamicVars))
DynVarsContext.displayName = 'DynVars'

/**
 * A `DynVarsProvider` allows providing your own dynamic variables for a
 * (sub)tree of components, overriding the globally passed-in dynamic variables.
 * Normally, application might want to just use the `useDynVars()` hook without
 * any `DynVarsProvider` of their own: this automatically default to providing
 * the dynamic variables as passed in by the server (if any).
 */
export const DynVarsProvider = DynVarsContext.Provider

/**
 * Returns the dynamic variables as passed in or overridden in a dynamic
 * variables context provider; {} if no variables were passed in.
 */
export const useDynVars = () => useContext(DynVarsContext)

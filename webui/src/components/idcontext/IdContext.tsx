// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'

const idPrefix = 'idcontext.'
const idSuffix = '-'
const digits = 8

// The DOM element identifier context; we initialize it to a default value that
// can be easily spotted for better trouble shooting. 
const idContext = React.createContext(`${idPrefix}${'x'.repeat(digits)}${idSuffix}`)
idContext.displayName = 'IdContext'

export interface IdContextProps {
    /** children inside a new link identifier context. */
    children: React.ReactNode
}

/**
 * `IdContext` provides a (new) unique DOM element identifier context (=prefix).
 * It allows to "namespace" DOM element identifiers in such situations where
 * React refs are not used for various reasons, keeping different views or
 * contexts within a React app separate in terms of their DOM element
 * identifiers.
 *
 * Components working directly with DOM element identifiers (instead of React
 * refs) need to construct their DOM element identifiers using the prefix
 * returned by the `useIdContext()` hook.
 */
export const IdContext = ({ children }: IdContextProps) => {
    return (
        <idContext.Provider
            value={`${idPrefix}${Math.floor(Math.random() * Math.pow(10, digits)).toString().padStart(digits, '0')}${idSuffix}`}
        >
            {children}
        </idContext.Provider>
    )
}

/**
 * useContextualId returns a new DOM element identifier based on an element
 * identifier prefix and the specified id parameter. The prefix is taken from
 * the nearest parent IdContext component.
 */
export const useContextualId = (id: string) => useContext(idContext) + id

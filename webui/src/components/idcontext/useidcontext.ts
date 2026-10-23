// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'

const idPrefix = 'idcontext.'
const idSuffix = '-'
const digits = 8

// The DOM element identifier context; we initialize it to a default value that
// can be easily spotted for better trouble shooting. 
export const idContext = React.createContext(`${idPrefix}${'x'.repeat(digits)}${idSuffix}`)
idContext.displayName = 'IdContext'

export const createId = () =>
    `${idPrefix}${Math.floor(Math.random() * 10 ** digits)
        .toString()
        .padStart(digits, '0')}${idSuffix}`

/**
 * useContextualId returns a new DOM element identifier based on an element
 * identifier prefix and the specified id parameter. The prefix is taken from
 * the nearest parent IdContext component.
 */
export const useContextualId = (id: string) => useContext(idContext) + id

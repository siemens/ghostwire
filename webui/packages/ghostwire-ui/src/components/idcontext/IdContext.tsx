// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'
import { createId, idContext } from './useidcontext'

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
    const [id] = useState(() =>createId())

    return (
        <idContext.Provider value={id}>
            {children}
        </idContext.Provider>
    )
}

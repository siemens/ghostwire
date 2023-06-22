// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

/**
 *  Hook monitoring the (address/url) location for changes and if the new
 *  location includes a "#" hash part it tries to scroll the DOM element having
 *  this id attribute value into view.
 *
 * @param scroller function taking a single DOM id string parameter and then
 * scrolling the DOM element with id into view.
 */
export const useScrollToHash = (scroller: (id: string) => void) => {

    const location = useLocation()

    useEffect(() => {
        const { hash } = location
        if (hash && scroller) {
            const id = hash.replace('#', '')
            scroller(id)
        }
    }, [location, scroller])
}

export default useScrollToHash

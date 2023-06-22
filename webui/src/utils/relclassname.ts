// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { NetworkInterface, nifId } from 'models/gw'


/**
 * Returns a (stable) CSS class name for a relation between two network
 * interfaces. Stable here refers to the returned class name always containing
 * the lexicographically earlier network interface before the later one, thus
 * returning the same class name regardless of the order in which the two
 * related network interfaces are passed in.
 *
 * As a special case, the second network interface object might be undefined
 * to handle the situation of relations to the outside ("external wires").
 *
 * @param domIdBase DOM element ID context base; used for scoping.
 * @param nifA first network interface object.
 * @param nifB optional second network interface object.
 */
export const relationClassName = (domIdBase: string, nifA: NetworkInterface, nifB?: NetworkInterface) => {
    const idA = nifId(nifA)
    if (!nifB) {
        return `${domIdBase}rel-${idA}`
    }
    const idB = nifId(nifB)
    if (idA.localeCompare(idB) <= 0) {
        return `${domIdBase}rel-${idA}-${idB}`
    }
    return `${domIdBase}rel-${idB}-${idA}`
}

/**
 * Returns a (stable) CSS class name for a relation between two network
 * interfaces, given only the DOM element IDs instead of network interface
 * objects.
 *
 * @param domIdBase 
 * @param nifIdA 
 * @param nifIdB 
 */
export const relationClassNameFromIds = (domIdBase: string, nifIdA: string, nifIdB?: string) => {
    const idA = nifIdA.startsWith(domIdBase) ? nifIdA.slice(domIdBase.length) : nifIdA
    if (!nifIdB) {
        return `${domIdBase}rel-${idA}`
    }
    const idB = nifIdB.startsWith(domIdBase) ? nifIdB.slice(domIdBase.length) : nifIdB
    if (idA.localeCompare(idB) <= 0) {
        return `${domIdBase}rel-${idA}-${idB}`
    }
    return `${domIdBase}rel-${idB}-${idA}`
}

export const isRelationClassName = (domIdBase: string, className: string) => {
    return className.startsWith(`${domIdBase}rel-`)
}

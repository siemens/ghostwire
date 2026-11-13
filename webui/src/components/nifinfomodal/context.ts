// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'
import { type NetworkInterface } from 'models/gw'

export const NifInfoModalContext = React.createContext<
    undefined | React.Dispatch<React.SetStateAction<NetworkInterface | undefined>>>(undefined)

/**
 * Returns a setter to specify the Nif to show information about in a modal
 * dialog.
 */
export const useNifInfoModal = () => {
    return useContext(NifInfoModalContext)
}

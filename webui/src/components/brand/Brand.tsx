// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useDynVars } from 'components/dynvars'

/**
 * Renders the brand name as a simple string: this is either the default
 * "Ghostwire" brand or an optional brand name override via the dynamic
 * variables passed to us by the Ghostwire service.
 */
export const Brand = () => {

    const { brand } = useDynVars()

    return brand ? String(brand) : 'Ghostwire'
}

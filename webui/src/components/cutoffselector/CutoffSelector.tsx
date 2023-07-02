// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { PrimitiveAtom, useAtom } from 'jotai'
import { MenuItem, Select, SelectChangeEvent } from '@mui/material'
import useId from 'hooks/id/id'


// Arbitrary high cutoff value
export const NEVER = 1e6


export interface CutoffSelectorProps {
    /** atom for storing the selected cutoff number. */
    atom: PrimitiveAtom<number>
    /** element name */
    element: string
    elements?: string
    /** minimal width in ems */
    em?: number
}

export const CutoffSelector = ({ atom, element, elements, em }: CutoffSelectorProps) => {

    const [cutoff, setCutoff] = useAtom(atom)
    const id = useId()

    // Yeah, very poor I18N.
    elements = elements || `${element}s`

    em = em >= 5 ? em : 5

    const handleChange = (event: SelectChangeEvent<number>) => {
        setCutoff(event.target.value as number)
    }

    return (
        <Select
            size="small"
            id={id}
            value={cutoff}
            onChange={handleChange}
            style={{ minWidth: `${em}em` }}
        >
            <MenuItem value={0}>always collapse</MenuItem>
            {[1, 2, 3, 4, 5, 10, 15, 20, 30].map(cutoff => (
                <MenuItem key={cutoff} value={cutoff}>
                    {cutoff !== 1 ? 'up to ' : ''}{cutoff} {cutoff === 1 ? element : elements}
                </MenuItem>
            ))}
            <MenuItem key={NEVER} value={NEVER}>show always</MenuItem>
        </Select>
    )
}

export default CutoffSelector

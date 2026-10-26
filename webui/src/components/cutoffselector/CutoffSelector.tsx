// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { type PrimitiveAtom, useAtom } from 'jotai'
import { MenuItem, Select, type SelectChangeEvent } from '@mui/material'
import useId from 'hooks/id/id'


// Arbitrary high cutoff value
export const NEVER = 1e6

const cutoffs = [1, 2, 3, 4, 5, 10, 15, 20, 30]


export interface CutoffSelectorProps {
    /** atom for storing the selected cutoff number. */
    atom: PrimitiveAtom<number>
    /** element name */
    element: string
    elements?: string
    /** minimal width in ems */
    em?: number
}

export const CutoffSelector = ({ atom, element, elements: elems, em: ems }: CutoffSelectorProps) => {

    const [cutoff, setCutoff] = useAtom(atom)
    const id = useId()

    // Yeah, very poor I18N.
    const elements = elems || `${element}s`

    const em = (ems && ems >= 5) ? ems : 5

    // Keep a stored value outside the predefined cutoffs selectable.
    const options = cutoff > 0 && cutoff !== NEVER && !cutoffs.includes(cutoff)
        ? [...cutoffs, cutoff].sort((a, b) => a - b)
        : cutoffs

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
            {options.map(n => (
                <MenuItem key={n} value={n}>
                    {n !== 1 ? 'up to ' : ''}{n} {n === 1 ? element : elements}
                </MenuItem>
            ))}
            <MenuItem key={NEVER} value={NEVER}>show always</MenuItem>
        </Select>
    )
}

export default CutoffSelector

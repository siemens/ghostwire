// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import { Checkbox, styled } from '@mui/material'
import CaptureCheckIcon from 'icons/CaptureCheck'
import { NetworkInterface, isOperational } from 'models/gw'
import React from 'react'

const NifChecker = styled(Checkbox)(() => ({
    borderRadius: '50px',
    maxWidth: '26px',
    padding: '3px 3px',
}))


export interface NifCheckboxProps {
    className?: string
    nif: NetworkInterface
    checked?: boolean
    onChange: (event: React.ChangeEvent<HTMLInputElement>, checked: boolean) => void
}

/**
 * `NifCheckbox` renders a checkbox for selecting individual network interfaces
 * for capture and displaying a shark fin instead of a check mark when selected.
 * If the specified network interface isn't operational, then the checkbox
 * cannot be selected.
 */
export const NifCheckbox = ({className, nif, checked, onChange}:NifCheckboxProps) => {
    return <NifChecker
        className={className}
        size="small"
        checkedIcon={<CaptureCheckIcon />}
        disabled={!isOperational(nif)}
        checked={checked}
        onChange={onChange}
    />
}

export default NifCheckbox

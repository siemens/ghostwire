// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { useMemo, useState } from 'react'

import { Box, IconButton, TextField, ToggleButton, ToggleButtonGroup, debounce } from '@mui/material'
import CaseIcon from 'icons/Case'
import RegexpIcon from 'icons/Regexp'
import { Clear } from '@mui/icons-material'
import { type FilterPattern } from './filter'

export interface FilterInputProps {
    /**
     * an optional filter pattern, including the matching options, such as
     * case-sensitive and interpretation of the pattern as a regexp instead of a
     * verbatim substring.
     */
    filterPattern?: FilterPattern
    /**
     * callback that is called whenever the filter pattern or its options
     * changes; this callback already is debounced.
     * 
     * @param fp filter patting including options.
     * @returns void
     */
    onChange: (fp: FilterPattern) => void
    /**
     * callback when the Enter key is pressed.
     * 
     * @returns void
     */
    onEnter?: () => void
    /** pattern change callback debounce wait in milliseconds. */
    debounceWait?: number
    /**
     * an optional React reference (object) that will be set to the text input's
     * HTML input field. 
     */
    focusRef?: React.RefObject<HTMLDivElement | null>
}

export const FilterInput = ({ filterPattern, onChange, debounceWait = 300, focusRef, onEnter }: FilterInputProps) => {

    const [pattern, setPattern] = useState('')
    const [filterOptions, setFilterOptions] = useState<string[]>([])

    const [prevFilterPattern, setPrevFilterPattern] = useState<FilterPattern>()
    if (filterPattern !== prevFilterPattern) {
        setPrevFilterPattern(filterPattern)
        setPattern(filterPattern?.pattern || '')
        setFilterOptions((filterPattern?.isCaseSensitive ? ["case"] : []).concat(filterPattern?.isRegexp ? ["regexp"] : []))
    }

    const debouncedOnChange = useMemo(
        () => debounce((pattern: string, options: string[]) => {
            if (!onChange) {
                return
            }
            onChange({
                pattern: pattern,
                isCaseSensitive: options.includes('case'),
                isRegexp: options.includes('regexp'),
            })
        }, debounceWait),
        [onChange, debounceWait])

    const handleInput = (event: React.ChangeEvent<HTMLInputElement>) => {
        const newPattern = event.target.value
        setPattern(newPattern)
        debouncedOnChange(newPattern, filterOptions)
    }

    const handleClear = () => {
        const newPattern = ''
        setPattern(newPattern)
        debouncedOnChange(newPattern, filterOptions)
    }

    const handleOptions = (_event: React.MouseEvent<HTMLElement>, newopts: string[]) => {
        setFilterOptions(newopts)
        debouncedOnChange(pattern, newopts)
    }

    const handleEnter = (event: React.KeyboardEvent) => {
        if (event.key !== 'Enter' || !onEnter) {
            return
        }
        onEnter()
    }

    // If the pattern is to be used as a regular expression, do a dry run in
    // order to determine whether the regexp pattern is valid or not. We later
    // use this to control the text input field's error indication.
    let regexpError = false
    if (filterOptions.includes('regexp')) {
        try {
            new RegExp(pattern)
        } catch {
            regexpError = true
        }
    }

    return <Box sx={{ display: "inline-flex", alignItems: "center", width: "100%" }}>
        <TextField
            placeholder="filter"
            value={pattern}
            size="small"
            variant="standard"
            sx={{ flexGrow: 1 }}
            inputRef={focusRef}
            error={regexpError}
            onChange={handleInput}
            onKeyDown={handleEnter}
            slotProps={{
                input: {
                    spellCheck: false,
                    endAdornment: <IconButton
                        sx={{ visibility: pattern && 'visible' || 'hidden' }}
                        onClick={handleClear}
                        size="small"
                    >
                        <Clear fontSize="small" />
                    </IconButton>
                },
            }}
        />
        <ToggleButtonGroup
            size="small"
            sx={{ pl: 1 }}
            onChange={handleOptions}
            value={filterOptions}
        >
            <ToggleButton value="case">
                <CaseIcon />
            </ToggleButton>
            <ToggleButton value="regexp">
                <RegexpIcon />
            </ToggleButton>
        </ToggleButtonGroup>
    </Box>
}

export default FilterInput

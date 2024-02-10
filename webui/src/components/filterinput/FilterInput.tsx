// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { useEffect, useMemo, useState } from 'react'

import { Box, IconButton, TextField, ToggleButton, ToggleButtonGroup, debounce } from '@mui/material'
import CaseIcon from 'icons/Case'
import RegexpIcon from 'icons/Regexp'
import { Clear } from '@mui/icons-material'

export interface FilterPattern {
    pattern: string
    isCaseSensitive: boolean
    isRegexp: boolean
}

export const getFilterFn = (fp: FilterPattern) => {
    if (!fp.isRegexp) {
        if (fp.isCaseSensitive) {
            return (s: string) => s.includes(fp.pattern)
        }
        const filter = fp.pattern.toLocaleLowerCase()
        return (s: string) => s.toLocaleLowerCase().includes(filter)
    }
    try {
        const re = new RegExp(fp.pattern, fp.isCaseSensitive ? "" : "i")
        return (s: string) => re.test(s)
    } catch (e) {
        return () => false
    }
}

export interface FilterInputProps {
    filterPattern?: FilterPattern
    onChange: (fp: FilterPattern) => void
    debounceWait?: number
}

export const FilterInput = ({ filterPattern, onChange, debounceWait }: FilterInputProps) => {

    debounceWait = debounceWait || 300

    const [pattern, setPattern] = useState('')
    const [filterOptions, setFilterOptions] = useState<string[]>([])

    useEffect(() => {
        setPattern(filterPattern?.pattern || '')
        setFilterOptions((filterPattern?.isCaseSensitive ? ["case"] : []).concat(filterPattern?.isRegexp ? ["regexp"] : []))
    }, [filterPattern])

    const onChangeHandler = (pattern: string, options: string[]) => {
        if (!onChange) {
            return
        }
        const isCaseSensitive = options.includes('case')
        const isRegexp = options.includes('regexp')
        const fp = {
            pattern: pattern,
            isCaseSensitive: isCaseSensitive,
            isRegexp: isRegexp,
        } as FilterPattern
        onChange(fp)
    }

    const debouncedOnChange = useMemo(
        () => debounce(onChangeHandler, debounceWait),
        [onChange])

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

    const handleOptions = (event: React.MouseEvent<HTMLElement>, newopts: string[]) => {
        setFilterOptions(newopts)
        debouncedOnChange(pattern, newopts)
    }

    // If the pattern is to be used as a regular expression, do a dry run in
    // order to determine whether the regexp pattern is valid or not. We later
    // use this to control the text input field's error indication.
    let regexpError = false
    if (filterOptions.includes('regexp')) {
        try {
            new RegExp(pattern)
        } catch (e) {
            regexpError = true
        }
    }

    return <Box sx={{ display: "inline-flex", alignItems: "center", width: "100%" }}>
        <TextField
            sx={{ flexGrow: 1 }}
            size="small"
            variant="standard"
            placeholder="filter"
            error={regexpError}
            onChange={handleInput}
            value={pattern}
            InputProps={{
                endAdornment: <IconButton
                    sx={{ visibility: pattern && 'visible' || 'hidden' }}
                    onClick={handleClear}
                >
                    <Clear fontSize="small" />
                </IconButton>
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

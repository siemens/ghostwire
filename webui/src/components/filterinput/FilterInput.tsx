// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'

import { Box, IconButton, TextField, ToggleButton, ToggleButtonGroup } from '@mui/material'
import CaseIcon from 'icons/Case'
import RegexpIcon from 'icons/Regexp'
import { Clear } from '@mui/icons-material'

export const FilterInput = () => {

    const [filter, setFilter] = useState('')
    const [filterOptions, setFilterOptions] = useState<string[]>(() => [])

    const handleInput = (event: React.ChangeEvent<HTMLInputElement>) => {
        setFilter(event.target.value)
    }

    const handleClear = () => {
        setFilter('')
    }

    const handleOptions = (event: React.MouseEvent<HTMLElement>, newopts: string[]) => {
        setFilterOptions(newopts)
    }

    return <Box sx={{ display: "inline-flex", alignItems: "center", width: "100%" }}>
        <TextField
            sx={{ flexGrow: 1 }}
            size="small"
            variant="standard"
            placeholder="filter"
            onChange={handleInput}
            value={filter}
            InputProps={{
                endAdornment: <IconButton
                    sx={{ visibility: filter && 'visible' || 'hidden'}}
                    onClick={handleClear}
                >
                    <Clear fontSize="small"/>
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
            <ToggleButton value="regex">
                <RegexpIcon />
            </ToggleButton>
        </ToggleButtonGroup>
    </Box>
}

export default FilterInput

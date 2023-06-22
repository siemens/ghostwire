// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useAtom } from 'jotai'

import { discoveryRefreshingAtom } from 'components/discovery'

import { IconButton, Tooltip } from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'

const RefreshButton = () => {

    const [, setRefreshing] = useAtom(discoveryRefreshingAtom)

    return (
        <Tooltip title="refresh">
            <IconButton color="inherit" size="small"
                onClick={() => setRefreshing(true)}
            >
                <RefreshIcon style={{ verticalAlign: 'middle' }} />
            </IconButton>
        </Tooltip>
    )

}

export default RefreshButton

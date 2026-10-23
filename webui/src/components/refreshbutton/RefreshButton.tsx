// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT


import { useAtom } from 'jotai'

import { discoveryRefreshingAtom } from 'components/discovery'

import { IconButton, Tooltip } from '@mui/material'
import RefreshIcon from '@mui/icons-material/Refresh'

/**
 * Renders a small refresh icon button that triggers an ad-hoc discovery
 * refresh when clicked; clicks while a refresh is already ongoing are ignored.
 */
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

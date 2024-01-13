// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useAtom } from 'jotai'

import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { Box, Typography } from '@mui/material'

import { useDiscovery } from 'components/discovery'
import { Ghost } from 'components/ghost'
import RefreshButton from 'components/refreshbutton'
import Metadata from 'components/metadata'
import { OpenPortTable } from 'components/openporttable/OpenPortTable'
import { showIpFamiliesAtom } from 'views/settings'

import OpenHouseIcon from 'icons/views/OpenHouse'

/**
 * Renders the discovered port forwardings and open ports from the host's
 * (=initial) network namespace.
 */
export const OpenHouse = React.forwardRef<HTMLDivElement, React.BaseHTMLAttributes<HTMLDivElement>>((props, ref) => {

    const [showIpFamilies] = useAtom(showIpFamiliesAtom)

    const discovery = useDiscovery()
    const netnses = Object.values(discovery.networkNamespaces)
    const netns = netnses.filter(netns => netns.isInitial).shift()

    return (
        <Box m={1}>
            {(netnses.length !== 0 &&
                <div ref={ref} /* so we can take a snapshot */>
                    <Metadata />
                    <Typography variant="body1" color="textSecondary">
                        <OpenHouseIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                        Open & Forwarding Host Ports
                    </Typography>
                    {netns &&
                        <OpenPortTable netns={netns} families={showIpFamilies} />
                    }
                </div>)
                || (<Ghost m={1}>
                    <Typography variant="body1" color="textSecondary" ref={ref}>
                        <InfoOutlinedIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                        nothing discovered yet, please refresh <RefreshButton />
                    </Typography>
                </Ghost>)}
        </Box>
    )
})
OpenHouse.displayName = "OpenHouse"

export default OpenHouse

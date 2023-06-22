// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useAtom } from 'jotai'

import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { Box, Typography } from '@mui/material'

import { useDiscovery } from 'components/discovery'
import { orderNetnsByContainees } from 'models/gw'
import { NetnsBreadboard } from 'components/netnsbreadboard'
import { showEmptyNetnsAtom, showIpFamiliesAtom, showLoopbackAtom } from 'views/settings'
import { Ghost } from 'components/ghost'
import RefreshButton from 'components/refreshbutton'
import Metadata from 'components/metadata'


/**
 * Renders the wiring of all discovered network namespaces. In case, nothing
 * was (yet) discovered, this component renders a short informational notice
 * informing users that they need to refresh the discovery in order to see
 * wires.
 */
export const NetnsWiring = React.forwardRef<HTMLDivElement, React.BaseHTMLAttributes<HTMLDivElement>>((props, ref) => {

    const [showLoopbacks] = useAtom(showLoopbackAtom)
    const [showEmptyNetns] = useAtom(showEmptyNetnsAtom)
    const [showIpFamilies] = useAtom(showIpFamiliesAtom)

    const discovery = useDiscovery()
    const netnses = Object.values(discovery.networkNamespaces)
        .sort(orderNetnsByContainees)

    return (
        <Box m={0} flex={1} overflow="auto">
            {(netnses.length !== 0 &&
                <div ref={ref} /* so we can take a snapshot */>
                    <Metadata />
                    <NetnsBreadboard
                        netns={netnses}
                        filterLo={!showLoopbacks}
                        filterEmpty={!showEmptyNetns}
                        families={showIpFamilies}
                    />
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

export default NetnsWiring

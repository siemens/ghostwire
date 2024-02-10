// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useAtom } from 'jotai'

import FilterAltIcon from '@mui/icons-material/FilterAlt'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { Box, Typography } from '@mui/material'

import { useDiscovery } from 'components/discovery'
import { orderNetnsByContainees } from 'models/gw'
import { NetnsBreadboard } from 'components/netnsbreadboard'
import { filterCaseSensitiveAtom, filterPatternAtom, filterRegexpAtom, showEmptyNetnsAtom, showIpFamiliesAtom, showLoopbackAtom } from 'views/settings'
import { Ghost } from 'components/ghost'
import RefreshButton from 'components/refreshbutton'
import Metadata from 'components/metadata'
import { getFilterFn } from 'components/filterinput'


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

    const [filterPattern] = useAtom(filterPatternAtom)
    const [filterCase] = useAtom(filterCaseSensitiveAtom)
    const [filterRegexp] = useAtom(filterRegexpAtom)
    const filterfn = getFilterFn({
        pattern: filterPattern,
        isCaseSensitive: filterCase,
        isRegexp: filterRegexp,
    })

    const discovery = useDiscovery()
    const orignetnses = Object.values(discovery.networkNamespaces)
    const netnses = orignetnses
        .filter(ns => {
            if (ns.containers.find(primcntee => filterfn(primcntee.name))) {
                return true
            }
            return ns.pods.find(pod => filterfn(pod.name))
        })
        .sort(orderNetnsByContainees)

    return (
        <Box m={0} flex={1} overflow="auto">
            {(netnses.length &&
                <div ref={ref} /* so we can take a snapshot */>
                    <Metadata />
                    <NetnsBreadboard
                        netns={netnses}
                        filterLo={!showLoopbacks}
                        filterEmpty={!showEmptyNetns}
                        families={showIpFamilies}
                    />
                </div>)
                || (orignetnses.length &&
                    <Typography m={1} variant="body1" color="textSecondary" ref={ref}>
                        <FilterAltIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                        no matches, please check the filter settings in the sidebar.
                    </Typography>)
                || (<Ghost m={1}>
                    <Typography variant="body1" color="textSecondary" ref={ref}>
                        <InfoOutlinedIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                        nothing discovered yet, please refresh <RefreshButton />
                    </Typography>
                </Ghost>)}
        </Box>
    )
})
NetnsWiring.displayName = "NetnsWiring"

export default NetnsWiring

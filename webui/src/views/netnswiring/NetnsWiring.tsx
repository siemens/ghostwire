// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useAtom } from 'jotai'

import FilterAltIcon from '@mui/icons-material/FilterAlt'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'
import { Box, Typography } from '@mui/material'

import { useDiscovery } from 'components/discovery'
import { emptyNonContainerNetns, netnsMatchesFilter, orderNetnsByContainees } from 'models/gw'
import { NetnsBreadboard } from 'components/netnsbreadboard'
import { filterCaseSensitiveAtom, filterPatternAtom, filterRegexpAtom, showEmptyNetnsAtom, showIpFamiliesAtom, showLoopbackAtom } from 'views/settings'
import { Ghost } from 'components/ghost'
import RefreshButton from 'components/refreshbutton'
import Metadata from 'components/metadata'
import { getFilterFn } from 'components/filterinput'
import { StripedNotice } from 'components/stripednotice'


/**
 * Renders the wiring of all discovered network namespaces. In case, nothing
 * was (yet) discovered, this component renders a short informational notice
 * informing users that they need to refresh the discovery in order to see
 * wires.
 */
export const NetnsWiring = React.forwardRef<HTMLDivElement, React.BaseHTMLAttributes<HTMLDivElement>>((_props, ref) => {

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
    // The candidate network namespaces for containee filter pattern matching:
    // the "show empty network namespaces" user setting has already been taken
    // into account here. These candidates also are the baseline deciding
    // whether an empty result is due to the filter pattern ("no matches") or
    // due to nothing having been discovered yet ("please refresh").
    const candidateNetnses = Object.values(discovery.networkNamespaces)
        .filter(netns => showEmptyNetns || !emptyNonContainerNetns(netns))
    // The network namespaces to actually render, additionally narrowed down by
    // the containee filter pattern; this is exactly what the breadboard will
    // show, without any further dropping downstream.
    const netnses = candidateNetnses
        .filter(ns => netnsMatchesFilter(ns, filterfn))
        .sort(orderNetnsByContainees)

    return (
        <Box sx={{ m: 0, flex: 1, overflow: 'auto' }}>
            {(netnses.length &&
                <div ref={ref} /* so we can take a snapshot */>
                    <Metadata />
                    {!!filterPattern &&
                        <StripedNotice>
                            <FilterAltIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                            filtering applied
                        </StripedNotice>}
                    <NetnsBreadboard
                        netns={netnses}
                        filterLo={!showLoopbacks}
                        families={showIpFamilies}
                    />
                </div>)
                || (candidateNetnses.length &&
                    <Typography sx={{ m: 1 }} variant="body1" color="textSecondary" ref={ref}>
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

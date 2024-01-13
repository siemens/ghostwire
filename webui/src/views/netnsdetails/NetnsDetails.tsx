// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useMatch, useNavigate } from 'react-router-dom'

import { useAtom } from 'jotai'

import { findBestMatch } from 'string-similarity'

import { Box, styled, Typography } from '@mui/material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'

import { useDiscovery } from 'components/discovery'
import { Containee, isPod, NetworkNamespace } from 'models/gw'
import { NetnsDetailCard } from 'components/netnsdetailcard'
import { showIpFamiliesAtom, showLoopbackAtom, showMACAtom } from 'views/settings'
import { Ghost } from 'components/ghost'
import { ContaineeBadge } from 'components/containeebadge'
import RefreshButton from 'components/refreshbutton'
import Metadata from 'components/metadata';


const SeeAlsoList = styled('ul')(({ theme }) => ({
    listStyle: 'none',
    listStylePosition: 'inside',

    '& li + li': {
        marginTop: theme.spacing(1),
    },
}))


/**
 * `NetnsDetails` renders the detailed information about a single network
 * namespace, where the network namespace is identified by either its namespace
 * identifier (inode number) or the name of one of the containees
 * attached to this network namespace. The identifier or name is passed as the
 * second path element in the current route, that is `/.../:slug` (with `:slug`
 * being either an inode number or an encoded URI component representing a
 * tenant name).
 *
 * Tenant names are matched using Dice's coefficient (sorry, Levenshtein), but
 * the best match is only accepted for a coefficient >= 0.7.
 *
 * If no sufficiently good match was found, then a list of less good matches is
 * rendered; users can click on one of these results in order to navigate to it,
 * replacing the current browser history element with the choosen match.
 *
 * The idea here is that user can view the communication configuration details
 * of a specific container even across container restarts that often might
 * result otherwise in the network namepace identifiers to change. Identifying
 * network namespaces indirectly be the name of one of its containees instead of
 * ever-changing inode numbers thus results in a much better UX.
 */
export const NetnsDetails = React.forwardRef<HTMLDivElement, React.BaseHTMLAttributes<HTMLDivElement>>((props, ref) => {
    const discovery = useDiscovery()

    const [showLoopbacks] = useAtom(showLoopbackAtom)
    const [showMAC] = useAtom(showMACAtom)
    const [families] = useAtom(showIpFamiliesAtom)

    // If the "slug" looks like an ID, try to match it to a network namespace
    // identifier first.
    const navigate = useNavigate()
    const match = useMatch('/:view/:slug')
    let netns: NetworkNamespace | undefined
    const netnsid = parseInt((match?.params as { [key: string]: string })['slug'])
    if (!isNaN(netnsid)) {
        netns = Object.values(discovery.networkNamespaces)
            .find(netns => netns.netnsid === netnsid)
    }
    // Either the slug didn't look like an namespace identifier number or we
    // didn't get a match, so let's try to match with a tenant name instead...
    let didyoumean: Containee[] = []
    if (!netns) {
        const tenantname = decodeURIComponent((match?.params as { [key: string]: string })['slug'])
        if (tenantname) {
            const containees = Object.values(discovery.networkNamespaces)
                // here, we really want to see ALL containees, including pod'ed
                // containers.
                .map(netns => (netns.containers as Containee[]).concat(netns.pods))
                .flat()
            if (containees.length) {
                // Determine the best match, but only accept it if it crosses a
                // certain threshold.
                const match = findBestMatch(tenantname, containees.map(tenant => tenant.name))
                if (match.bestMatch.rating >= 0.8) {
                    const bestMatch = containees[match.bestMatchIndex]
                    netns = isPod(bestMatch) ? bestMatch.containers[0].netns
                        : bestMatch.netns
                } else {
                    didyoumean = match.ratings
                        .filter((rat: { rating: number }) => rat.rating >= 0.2)
                        .map((rat: { target: string }) =>
                            containees.find(containee => containee.name === rat.target)
                        ) as Containee[]
                }
            }
        }
    }

    // Navigate to a specific containee on badge clicking ... "badge", not
    // "binge".
    const onContaineeClick = (containee: Containee) => {
        navigate(`/${(match?.params as { [key: string]: string })['view']}/${encodeURIComponent(containee.name)}`)
    }

    return (netns &&
        <Box m={0} flex={1} overflow="auto">
            <div ref={ref} /* so we can take a snapshot */>
                <Metadata />
                <Box m={1}>
                    <NetnsDetailCard
                        netns={netns}
                        filterLo={!showLoopbacks}
                        filterMAC={!showMAC}
                        families={families}
                        canMinimize
                    />
                </Box>
            </div>
        </Box >)
        || (<Ghost m={1}>
            <div ref={ref}>
                <Typography variant="body1" color="textSecondary" paragraph>
                    <InfoOutlinedIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                    unknown network namespace identifier (inode number),
                    pod, container, or stand-alone process name.
                </Typography>
                {(didyoumean.length &&
                    <Typography variant="body1" color="textSecondary">
                        Did you mean…?
                        <SeeAlsoList>
                            {didyoumean.map(containee =>
                                <li key={containee.name}>
                                    <ContaineeBadge
                                        containee={containee}
                                        button
                                        onClick={onContaineeClick}
                                    />
                                </li>
                            )}
                        </SeeAlsoList>
                    </Typography>) || null}
                <Typography variant="body1" color="textSecondary" paragraph>
                    Do you want to refresh? <RefreshButton />
                </Typography>
            </div>
        </Ghost>)
})
NetnsDetails.displayName = "NetnsDetails"

export default NetnsDetails

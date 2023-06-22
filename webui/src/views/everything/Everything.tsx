// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useLocation } from 'react-router-dom'

import { useAtom } from 'jotai'

import { Box, styled, Typography } from '@mui/material'
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined'

import { CardTray } from 'components/cardtray'
import { NetnsDetailCard } from 'components/netnsdetailcard'

import { useDiscovery } from 'components/discovery'
import { emptyNetns, isProject, orderNetnsByContainees, sortedNetnsProjects } from 'models/gw'
import { showEmptyNetnsAtom, showIpFamiliesAtom, showLoopbackAtom, showMACAtom } from 'views/settings'
import { Ghost } from 'components/ghost'
import RefreshButton from 'components/refreshbutton'

import { createMuiShadow } from 'utils/shadow'
import ProjectCard from 'components/projectcard/ProjectCard'
import Metadata from 'components/metadata'


const MarkableNetnsDetailCard = styled(NetnsDetailCard)(({ theme }) => ({
    '&.highlight': {
        boxShadow: createMuiShadow(theme.palette.routing.selected.dark, 0, 1, 5, 0, 0, 2, 2, 0, 0, 3, 1, -2),
        border: `1px solid ${theme.palette.routing.selected.light}`,
    },
    '&.normal': {
        border: `1px solid ${theme.palette.background.paper}`,
    },
}))


/**
 * Renders a full, detailed view of all discovered network namespaces.
 */
export const Everything = React.forwardRef<HTMLDivElement, React.BaseHTMLAttributes<HTMLDivElement>>((props, ref) => {
    const [showLoopbacks] = useAtom(showLoopbackAtom)
    const [showEmptyNetns] = useAtom(showEmptyNetnsAtom)
    const [showMAC] = useAtom(showMACAtom)
    const [families] = useAtom(showIpFamiliesAtom)

    // Did the user navigate to a specific network namespace...?
    const location = useLocation()
    const locmatch = location && location.hash.match(/^#.*-netns-(\d+)$/)
    const netnsid = (locmatch && parseInt(locmatch[1])) || 0

    const discovery = useDiscovery()
    const netnses = Object.values(discovery.networkNamespaces)
        .filter(netns => showEmptyNetns || !emptyNetns(netns))
        .sort(orderNetnsByContainees)

    const netnsesAndProjs = sortedNetnsProjects(netnses)

    return (netnses.length !== 0 &&
        <Box m={0} flex={1} overflow="auto">
            <div ref={ref} /* so we can take a snapshot */>
                <Metadata />
                <CardTray>
                    {netnsesAndProjs.map(netnsOrProj => {
                        if (isProject(netnsOrProj)) {
                            return <ProjectCard
                                key={netnsOrProj.name}
                                project={netnsOrProj}
                            >
                                {Object.values(netnsOrProj.netnses)
                                    .sort(orderNetnsByContainees)
                                    .map(netns =>
                                        <MarkableNetnsDetailCard
                                            key={netns.netnsid}
                                            className={netns.netnsid === netnsid ? 'highlight' : 'normal'}
                                            netns={netns}
                                            canMaximize
                                            filterLo={!showLoopbacks}
                                            filterMAC={!showMAC}
                                            families={families}
                                        />)
                                }
                            </ProjectCard>
                        } else {
                            return <MarkableNetnsDetailCard
                                key={netnsOrProj.netnsid}
                                className={netnsOrProj.netnsid === netnsid ? 'highlight' : 'normal'}
                                netns={netnsOrProj}
                                canMaximize
                                filterLo={!showLoopbacks}
                                filterMAC={!showMAC}
                                families={families}
                            />
                        }
                    })}
                </CardTray>
            </div>
        </Box >)
        || (<Ghost m={1}>
            <Typography variant="body1" color="textSecondary" ref={ref}>
                <InfoOutlinedIcon color="inherit" style={{ verticalAlign: 'middle' }} />&nbsp;
                nothing discovered yet, please refresh <RefreshButton />
            </Typography>
        </Ghost>)
})

export default Everything

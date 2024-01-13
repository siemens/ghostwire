// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { useNavigate, useLocation, useMatch } from 'react-router-dom'

import { styled } from '@mui/material'

import { AddressFamilySet, PrimitiveContainee, emptyNetns, NetworkInterface, NetworkNamespace, NetworkNamespaces, nifId, isProject, sortedNetnsProjects, orderNetnsByContainees, netnsId } from 'models/gw'
import { Breadboard } from 'components/breadboard/Breadboard'
import { CardTray } from 'components/cardtray'
import { NetnsPlainCard } from 'components/netnsplaincard'
import { useContextualId } from 'components/idcontext'
import { scrollIdIntoView } from 'utils'
import { createMuiShadow } from 'utils/shadow'
import ProjectCard from 'components/projectcard/ProjectCard'

const AccentableNetnsCard = styled(NetnsPlainCard)(({ theme }) => ({
    '&.normal': {
        border: `1px solid ${theme.palette.background.paper}`,
    },
    '&.highlight': {
        boxShadow: createMuiShadow(theme.palette.routing.selected.dark, 0, 1, 5, 0, 0, 2, 2, 0, 0, 3, 1, -2),
        border: `1px solid ${theme.palette.routing.selected.light}`,
    },
}))


export type NetworkNamespaceProp = NetworkNamespaces | NetworkNamespace[] | NetworkNamespace

export interface NetnsBreadboardProps {
    /** 
     * network namespaces for which to generate, layout and show the wiring.
     * The `Breadboard` component accepts not only the usualy discovery
     * network namespaces map, but also an array of network namespaces or even
     * a single network namespace.
     */
    netns?: NetworkNamespaceProp
    /** if `true`, then hides loopback interfaces. */
    filterLo?: boolean
    /** 
     * if `true`, then hides network namespaces with only loopback interfaces.
     */
    filterEmpty?: boolean
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
}

/**
 * `NetnsBreadboard` renders a the specified network namespaces with their
 * network interfaces and wiring. This surely is the trademark visual feature
 * of the Ghostwire UI.
 *
 * This breadboard implements navigation between related network interfaces,
 * so basically tracing the wires. User can click on network interface badges
 * which have related network interfaces: these appear as buttons with rounded
 * corners as opposed to plain badges with rectangular edges. Navigation
 * scrolls the user-selected related network interface into view.
 *
 * This breadboard also implements navigation ("zooming in") to a particular
 * network namespace or a specific containee.
 */
export const NetnsBreadboard = ({ netns, filterLo, filterEmpty, families }: NetnsBreadboardProps) => {
    const navigate = useNavigate()
    const match = useMatch('/:base')

    const domIdBase = useContextualId('')

    // Did the user navigate to a specific network namespace...?
    const location = useLocation()
    const locmatch = location && location.hash.match(/^#.*-netns-(\d+)$/)
    const netnsid = (locmatch && parseInt(locmatch[1])) || 0

    // To start with, bring the specified network namespace(s) into our
    // canonical form of an array of network namespaces.
    let netnses: NetworkNamespace[] = []
    if (Array.isArray(netns)) {
        netnses = [...netns] // ...because we might modify the array lateron.
    } else if (netns && 'netnsid' in netns) {
        netnses = [netns]
    } else if (netns) {
        netnses = Object.values(netns)
    }
    // Drop all namespaces with only a lonely loopback network interface, to
    // reduce noise. And yes, we even accept a network namespace without a
    // loopback network interface, but a different network interface.
    if (filterEmpty) {
        netnses = netnses.filter(netnsOrProj => isProject(netnsOrProj) || !emptyNetns(netnsOrProj as NetworkNamespace))
    }

    // Navigate to another (related) network interface within the same
    // breadbord. Hmm, virtual world follows physical world, and we even have
    // different wire colors! :D
    const handleNavigation = (nif: NetworkInterface) => {
        scrollIdIntoView(domIdBase + nifId(nif))
    }

    const handleNetnsZoom = (netns: NetworkNamespace, fragment?: string) => {
        let base = `/${match?.params['base']}/${netns.netnsid}`
        if (fragment) {
            base += `#${domIdBase}${netnsId(netns)}-${fragment}`
        }
        navigate(base)
    }

    // When the user clicks on the zoom button of a containee badge, we switch
    // to a detail view of exactly only the network namespace of this
    // containee, yet within the same URL base.
    const handleContaineeZoom = (containee: PrimitiveContainee) => {
        navigate(`/${match?.params['base']}/${encodeURIComponent(containee.name)}`)
    }

    const netnsesAndProjs = sortedNetnsProjects(netnses)

    // With many things prepared before, it's now a piece of cake we can have
    // and also eat to render the breadboard with the network namespaces with
    // network interfaces and their wiring. Get Wiring Done!
    return (
        <Breadboard netns={netnses}>
            <CardTray animate>
                {netnsesAndProjs.map(netnsOrProj => {
                    if (isProject(netnsOrProj)) {
                        return <ProjectCard
                            key={netnsOrProj.name}
                            project={netnsOrProj}
                        >
                            {Object.values(netnsOrProj.netnses)
                                .sort(orderNetnsByContainees)
                                .map(netns =>
                                    <AccentableNetnsCard
                                        key={netns.netnsid}
                                        className={netns.netnsid === netnsid ? 'highlight' : 'normal'}
                                        netns={loFilteredNetns(netns, filterLo || false)}
                                        families={families}
                                        onNavigation={handleNavigation}
                                        onNetnsZoom={handleNetnsZoom}
                                        onContaineeZoom={handleContaineeZoom}
                                    />)
                            }
                        </ProjectCard>
                    } else {
                        return <AccentableNetnsCard
                            key={netnsOrProj.netnsid}
                            className={netnsOrProj.netnsid === netnsid ? 'highlight' : 'normal'}
                            netns={loFilteredNetns(netnsOrProj, filterLo || false)}
                            families={families}
                            onNavigation={handleNavigation}
                            onNetnsZoom={handleNetnsZoom}
                            onContaineeZoom={handleContaineeZoom}
                        />
                    }
                })}
            </CardTray>
        </Breadboard>
    )
}

// Returns a network namespace with the lo interface filtered out, if told so.
// Otherwise returns the original network namespace.
const loFilteredNetns = (netns: NetworkNamespace, filterLo: boolean) => {
    if (!filterLo) {
        return netns
    }
    const newnetns = { ...netns, nifs: { ...netns.nifs } }
    const lo = Object.keys(newnetns.nifs).find(nifidx => newnetns.nifs[nifidx as unknown as number].name === 'lo')
    if (lo) {
        delete newnetns.nifs[lo as unknown as number]
    }
    return newnetns
}

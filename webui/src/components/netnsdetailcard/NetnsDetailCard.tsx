// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Link, LinkProps, useNavigate, useMatch } from 'react-router-dom'

import { IconButton, Paper, styled, Tooltip } from '@mui/material'
import FullscreenIcon from '@mui/icons-material/Fullscreen'
import FullscreenExitIcon from '@mui/icons-material/FullscreenExit'

import { CardSection } from 'components/cardsection'
import { ContaineeBadge } from 'components/containeebadge'
import { AddressFamily, AddressFamilySet, containeeKey, isPod, netnsId, NetworkInterface, NetworkNamespace, nifId, orderRoutes, routeKey, containeesOfNetns, RouteTableLocal, isContainer, isSandbox, sortContaineesByName, Containee, Container } from 'models/gw'
import { NifTree } from 'components/niftree'
import { Route } from 'components/route'
import { useContextualId } from 'components/idcontext'
import { DetailedContainees } from 'components/detailedcontainees'
import { TransportPortTable } from 'components/transportporttable'
import { scrollIdIntoView } from 'utils'
import { useAtom } from 'jotai'
import { containeesCutoffAtom, forwardedPortsCutoffAtom, neighborhoodCutoffAtom, nifsCutoffAtom, portsCutoffAtom, routesCutoffAtom, showMultiBroadcastRoutesAtom, showNamespaceIdsAtom, showSandboxesAtom } from 'views/settings'
import { Neighborhood } from 'components/neighborhood'
import { neighborhoodServiceContainerCount, neighborhoodServices } from 'utils/neighborhood'
import { ForwardPortTable } from 'components/forwardporttable/ForwardPortTable'


const NetnsPaper = styled(Paper)(({ theme }) => ({
    // All information on a paper (card) for a network namespace gets the
    // usual inner padding between information and the paper edges.
    padding: theme.spacing(2),
}))

// Style the "containees" )such as containers and stand-alone processes)
// which are listed along the top edge of any network namespace card.
const Containees = styled('div')(({ theme }) => ({
    position: 'relative',
    left: -theme.spacing(1),
    marginBottom: theme.spacing(2),
    display: 'flex',
    flexWrap: 'wrap',
    // Since CSS flex boxes don't offer "collapsible margins", we have to
    // emulate them by offsetting the margins by half the desired spacing
    // between the flex items (the containees). See also:
    // https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Flexible_Box_Layout/Mastering_Wrapping_of_Flex_Items
    margin: -theme.spacing(1),
    // For correctly spacing the list of containees, we assume that the
    // BoxedBadge components are immediate children of the containees
    // "container"; thus, we don't care about the specific element type, but
    // simply apply to the immediate children.
    '& > *': {
        flex: '0 0 auto',
        margin: theme.spacing(1),
        marginBottom: 0,
    }
}))

const NetnsInfo = styled('div')(({ theme }) => ({
    marginBottom: theme.spacing(2),
}))

const NetnsDescription = styled('span')(({ theme }) => ({
    color: theme.palette.text.secondary,
}))

const NetnsID = styled('span')(() => ({
}))

const SingleRoute = styled('div')(({ theme }) => ({
    // display individual routes as blocks with hanging indentation in case
    // the route information needs line breaks. Since CSS still has made
    // "hanging" not official -- probably for reasons beyond technical,
    // considering how the styling would read out loud -- we emulate hanging
    // indentation using padding, except for the first line.
    display: 'block',
    paddingLeft: '1.5em',
    textIndent: '-1.5em',

    '& + &': {
        marginTop: theme.spacing(0.5),
    }
}))


const MaxMinIconButton = (styled(IconButton)<LinkProps>(({ theme }) => ({
    float: 'right',
    // Move the right-floated button partly back into the padding, as the
    // button has a large "corona" and we thus get a more pleasing visual
    // alignment, especially with the tentant (boxed entity) badges along
    // the top of the network namespace card.
    position: 'relative',
    top: -theme.spacing(1),
    left: theme.spacing(1),
})) as unknown) as typeof IconButton


export interface NetnsDetailCardProps {
    /** 
     * network namespace object (with tons of information thanks to the
     * Ghostwire discovery engine). 
     */
    netns: NetworkNamespace
    /** if `true`, then hides loopback interfaces. */
    filterLo?: boolean
    /** if `true`, then hides MAC layer addresses. */
    filterMAC?: boolean
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** 
     * shows a maximize button which navigates/links to a single namespace
     * view (within the current route).
     */
    canMaximize?: boolean
    /**
     * shows a minimize button.
     */
    canMinimize?: boolean
    /** CSS class name(s). */
    className?: string
}

/**
 * Component `NetnsDetailCard` renders detailed information about a network
 * namespace in form of a card (a piece of paper). It shows the following
 * details in form of separate sections:
 *
 * - contained entities, such as containers and stand-alone processes.
 * - listening and connected transport ports.
 * - routing entries. 
 * - (hierarchical) list of network interfaces: bridge-enslaved network
 *   interfaces as well as macvlan network interfaces are shown in a sub-level
 *   of their master/macvlan network interfaces.
 *
 * Each section is collapsible and expandible by users. The initial expansion or
 * collapsion state is determined on the basis of section-specific "cutoff"
 * configuration settings. As soon as the threshold for a cutoff value gets
 * crossed, a section will start out as collapsed. This allows users to avoid
 * having to scroll over tall sections in which they might not be interested. A
 * typical example is a busy transport port section of the initial network
 * namespace.
 */
export const NetnsDetailCard = ({ netns, filterLo, filterMAC, families, canMaximize, canMinimize, className }: NetnsDetailCardProps) => {
    const navigate = useNavigate()
    const match1 = useMatch('/:base')
    const match2 = useMatch('/:base/:detail')
    const match = (match1 || match2) ? { ...match1, ...match2 } : null // groan, react router dom v6 still broken.

    const domIdBase = useContextualId('')
    const netnsid = domIdBase + netnsId(netns)

    families = families || [AddressFamily.IPv4, AddressFamily.IPv6]

    const [containeesCutoff] = useAtom(containeesCutoffAtom)
    const [neighborhoodCutoff] = useAtom(neighborhoodCutoffAtom)
    const [forwardedPortsCutoff] = useAtom(forwardedPortsCutoffAtom)
    const [portsCutoff] = useAtom(portsCutoffAtom)
    const [routesCutoff] = useAtom(routesCutoffAtom)
    const [nifsCutoff] = useAtom(nifsCutoffAtom)
    const [showSandboxes] = useAtom(showSandboxesAtom)
    const [showNamespaceIds] = useAtom(showNamespaceIdsAtom)
    const [showMultiBroadcastRoutes] = useAtom(showMultiBroadcastRoutesAtom)

    // Get the containees to be shown in the details section; please note that
    // we don't use this result for displaying the containee badges, as these
    // are two different use cases: one for (almost) all containee details, the
    // other to show a high-level view on what's attached to this network
    // namespace.
    const containees = netns.containers
        .filter(containee => !isSandbox(containee))
        .filter(containee => !isContainer(containee) || !containee.sandbox || showSandboxes)
    const containeesCutoffState = containees.length > containeesCutoff ? 'collapse' : 'expand'

    // Get the names of the services/containers reachable in the neighborhood,
    // that is, via the attached networks.
    const somecontainer = netns.containers.filter(cntr => isContainer(cntr)).pop() as Container
    const neighborServices = neighborhoodServices(somecontainer)
    const neighborhoodCutoffState = neighborhoodServiceContainerCount(neighborServices) > neighborhoodCutoff ? 'collapse' : 'expand'

    let fwportrows = 0
    const forwardedPortsCutoffState = (netns.forwardedPorts.find(port => {
        if (families.includes(port.address.family)) {
            fwportrows += port.users.length
            // Make find terminate its "search" as soon as we cross the cutoff.
            // Otherwise, keep on summing up...
            return fwportrows > forwardedPortsCutoff
        }
        // will not be shown, so keep find searching on...
        return false
    })) ? 'collapse' : 'expand'

    // Decide whether we'll cross the cutoff for expanding the transport port
    // table or not. We want to avoid all the filtering, sorting and mapping
    // here as this being done only in the transport port table component. Here,
    // it sufficies to find out if we're above the threshold or not. And here we
    // take "find" literally: we slightly mis-appropriate the find() array
    // operation to calculate the sum of transport table rows until we counted
    // all elements or until we cross the cutoff point. Please note that we have
    // to keep in mind that multiple processes might use the same port, not just
    // a single one.
    let portrows = 0
    const portsCutoffState = (netns.transportPorts.find(port => {
        if (families.includes(port.localAddress.family)) {
            portrows += port.users.length
            // Make find terminate its "search" as soon as we cross the cutoff.
            // Otherwise, keep on summing up...
            return portrows > portsCutoff
        }
        // will not be shown, so keep find searching on...
        return false
    })) ? 'collapse' : 'expand'

    // Nota bene: we don't show routes from the "local" table, so we filter it
    // out here in order to later correctly calculate the number of routes to be
    // rendered ... as needed to determine the collapsed/expanded state of the
    // route section.
    const routes = netns.routes
        .filter(route =>
            families.includes(route.family) && (showMultiBroadcastRoutes || route.table !== RouteTableLocal))
        .sort(orderRoutes)
    
    // Tiny helper to derive section component keys based on current network
    // namespace. This ensures that DOM update works correctly due to changing
    // section keys and avoids section header collapse/expand state not getting
    // correctly updated because the section header keeps using its old state.
    const key = (sect: string) => `${netns.netnsid}-${sect}`

    // When the user clicks on a (related) network interface button badge,
    // then navigate to this related interface, either by scrolling it into
    // view (in all/total view) or by switching the route to the detail which
    // contains the related network interface.
    const handleNifNavigation = (nif: NetworkInterface) => {
        if (!match) {
            return
        }
        if (match.params['detail']) {
            // change route from existing detail view to new detail view.
            navigate(`/${match.params['base']}/${nif.netns.netnsid}`)
        }
        // scroll within the overall view.
        scrollIdIntoView(domIdBase + nifId(nif))
    }

    // User clicked on the outgoing network interface of a route, so scroll
    // that network interface into view, if necessary.
    const handleRouteNavigation = (nif: NetworkInterface) => {
        scrollIdIntoView(domIdBase + nifId(nif))
    }

    // When the user clicks on some containee button badge that is listed as a
    // reference in one of the sections, then navigate to it. We have to
    // differentiate between navigating within the full view, which is actually
    // scrollinge into view, and navigating from one detail view to another
    // detail view, where we change the route.
    const handleContaineeNavigation = (containee: Containee) => {
        if (!match) {
            return
        }
        if (match.params['detail']) {
            // change route from existing detail view to new detail view.
            navigate(`/${match.params['base']}/${encodeURIComponent(containee.name)}`)
        }
        // scroll within the overall view.
        scrollIdIntoView(domIdBase + netnsId(
            isPod(containee) ? containee.containers[0].netns
                : containee.netns))
    }

    // When the user clicks on the containee button badges at the top of card as
    // opposed to referenced containee buttons in one of the sections, switch to
    // the corresponding detail view. This is never triggered while in a single
    // detail view.
    const handleMaximize = (containee: Containee) => {
        if (!match) {
            return
        }
        navigate(`/${match.params['base']}/${encodeURIComponent(containee.name)}`)
    }

    return (
        <NetnsPaper className={className}>
            {/* place the ID as an "anchor" at the *top* of the Paper component. */}
            <span id={netnsid} />

            {/* if needed, show a maximize (zoom in) button or a minimize (zoom out) button */}
            {canMaximize && match.params['base'] &&
                <Tooltip title="show only this network namespace">
                    <MaxMinIconButton
                        component={Link}
                        to={`/${match.params['base']}/${netns.netnsid}`}
                        size="large">
                        <FullscreenIcon />
                    </MaxMinIconButton>
                </Tooltip>}
            {canMinimize && match.params['base'] &&
                <Tooltip title="back to overall view">
                    <MaxMinIconButton
                        component={Link}
                        to={`/${match.params['base']}`}
                        size="large">
                        <FullscreenExitIcon />
                    </MaxMinIconButton>
                </Tooltip>}

            {/*
              * render (a series of) containee badge(s); please note that this
              * will render pods in place of individual containers, so not too
              * many details here.
              */}
            <Containees>
                {containeesOfNetns(netns)
                    .sort(sortContaineesByName)
                    .map(containee =>
                        <ContaineeBadge
                            containee={containee}
                            key={containeeKey(containee)}
                            capture
                            button={canMaximize}
                            angled={canMaximize}
                            endIcon={canMaximize && <FullscreenIcon />}
                            onClick={canMaximize && (() => handleMaximize(containee))}
                        />
                    )}
            </Containees>

            {showNamespaceIds &&
                <NetnsInfo>
                    <NetnsDescription>network namespace ID:</NetnsDescription>
                    {' '}<NetnsID>{netns.netnsid}</NetnsID>
                </NetnsInfo>
            }

            {/* details of containees, such as their DNS configuration */}
            {(containees.length &&
                <CardSection
                    key={key('containees')}
                    caption="containees"
                    collapsible={containeesCutoffState}
                    fragment={netnsid+'-containees'}
                >
                    <DetailedContainees
                        netns={netns}
                        families={families}
                        showSandbox={showSandboxes}
                    />
                </CardSection>) || ''}

            {/* neighborhood service names details -- hide if no neighbor services */}
            {neighborServices.length > 0 &&
                <CardSection
                    key={key('neighborhood')}
                    caption="neighborhood services (host-internal)"
                    collapsible={neighborhoodCutoffState}
                    fragment={netnsid+'-neighborhood'}
                >
                    <Neighborhood services={neighborServices} seenby={netns.containers} />
                </CardSection>
            }

            {/* forwarded transport port details ("docker ps totally drugged") */}
            <CardSection
                key={key('forwardedports')}
                caption="port forwarding"
                collapsible={forwardedPortsCutoffState}
                fragment={netnsid+'-fwports'}
                >
                <ForwardPortTable hideEmpty netns={netns} families={families} />
            </CardSection>

            {/* transport port details ("netstat/ss on steriods") */}
            <CardSection
                key={key('ports')}
                caption="transport"
                collapsible={portsCutoffState}
                fragment={netnsid+'-ports'}
                >
                <TransportPortTable hideEmpty netns={netns} families={families} />
            </CardSection>

            {/* routing details */}
            <CardSection
                key={key('routes')}
                caption="routing"
                collapsible={routes.length > routesCutoff ? 'collapse' : 'expand'}
                fragment={netnsid+'-routes'}
                >
                {routes.map(rt =>
                    <SingleRoute key={routeKey(rt)}>
                        <Route route={rt} onNavigation={handleRouteNavigation} />
                    </SingleRoute>
                )}
            </CardSection>

            {/* network interface details */}
            <CardSection
                key={key('nifs')}
                caption="network interfaces"
                collapsible={Object.keys(netns.nifs).length > nifsCutoff ? 'collapse' : 'expand'}
                fragment={netnsid+'-nifs'}
            >
                <NifTree
                    netns={netns}
                    filterLo={filterLo}
                    filterMAC={filterMAC}
                    families={families}
                    onNavigation={handleNifNavigation}
                    onContaineeNavigation={handleContaineeNavigation}
                />
            </CardSection>
        </NetnsPaper>
    )
}

export default NetnsDetailCard

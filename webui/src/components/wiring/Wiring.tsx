// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { ReactElement, useLayoutEffect, useState } from 'react'
import clsx from 'clsx'

import { styled, useTheme } from '@mui/material'

import { useContextualId } from 'components/idcontext'
import { isRelationClassName, relationClassNameFromIds } from 'utils/relclassname'


/**
 * High-level description of a (virtual) wire, either between two network
 * interfaces, or an outgoing wire from a single network interface to reality
 * (which must be somewhere outside here).
 */
export interface Wire {
    /** 
     * kind (type) of network interfaces, such as "veth", "macvlan", or "" for
     * wires going to the outside.
     */
    kind: string
    /** is the operational status of this wire down? */
    operStateDown: boolean
    /** First or outgoing network interface DOM identifier. */
    nif1Id: string
    /** 
     * Optionally second network interface DOM identifier, required for
     * point-to-point connections, including multipoint-to-point connection.
     */
    nif2Id?: string
}

/**
 * Layouted wire which has been assigned to a specific swim-lane and the start
 * and endpoints have been resolved.
 */
interface SwimlaneWire {
    /** DOM element representing a first network interface. */
    nif1Element: HTMLElement
    /** 
     * DOM element representing a second network interface, the plurality of
     * which ... *plonk* 
     */
    nif2Element: HTMLElement
    /** 
     * kind (type) of network interfaces, such as "veth", "macvlan", or "" for
     * wires going to the outside.
     */
    kind: string
    /** 
     * (vertical) length of wire, based on the layout of the network interface
     * DOM elements. 
     */
    len: number
    /** upper Y position of wire, where it starts. */
    from: number
    /** upper X inset of wire start. */
    fromInset: number
    /** lower Y position of wire, where it ends. */
    to: number
    /** lower Y inset of wire end. */
    toInset: number
    /** 
     * swim-lane number (starting from 0) where the vertical segment of the wire
     * should be placed into.
     */
    lane: number,
    /**
     * if true, the wire's direction is from nif2Element to nif1Element as
     * opposed to basic orientation running from nif1Element to nif2Element.
     * Setting reverse indicates that wire adornments that correspond with the
     * direction of the wire need to switch places between the first and second
     * network interface.
     */
    reverse: boolean
    /** TODO: tooltip to associate with this wire. */
    tooltip: string
    /** 
     * when one or both network interfaces are down, the wire is considered to
     * be down, too.
     */
    operStateDown: boolean
}

/** Simple rectangle description. */
interface Rect {
    top: number
    left: number
    bottom: number
    right: number
}

/**
 * Calculate the bounding rect of one DOM Element relative to some other, outer,
 * element's rect position and extension.
 *
 * @param domEl the element those bound rect we want to know, relative to the
 * specified reference rectangle.
 * @param refRect some out element which represents the reference frame. 
 */
const nifBoundingRect = (domEl: HTMLElement, refRect: DOMRect): Rect => {
    const nifRect = domEl.getBoundingClientRect()
    return {
        top: nifRect.top - refRect.top,
        left: nifRect.left - refRect.left,
        bottom: nifRect.bottom - refRect.top,
        right: nifRect.right - refRect.left,
    } as Rect
}

/**
 * Layout all (two-ended) wires into swim lanes.
 *
 * @param wires high-level wire information.
 * @param nifContainerDomId DOM element identifier of container with network
 * interface badges.
 * 
 * @returns [layouted wires, number of swin-laned required]
 */
const layoutWires = (wires: Wire[], nifContainerDomId: string): [SwimlaneWire[], number] => {
    // No wires yet in any swim-lanes.
    let lanes: (SwimlaneWire[])[] = []
    const refPosition = document.getElementById(nifContainerDomId).getBoundingClientRect()
    const refWidth = refPosition.right - refPosition.left
    // Determine the layout information for each individual wire.
    const wiresInLanes = wires.map(w => {
        const swr = {
            nif1Element: document.getElementById(w.nif1Id),
            nif2Element: document.getElementById(w.nif2Id),
            kind: w.kind,
            operStateDown: w.operStateDown,
        } as SwimlaneWire
        if (!swr.nif1Element || !swr.nif2Element) {
            return null
        }
        // Get the bounding boxes of the network interface badges, relative to
        // some outer DOM container.
        const nif1Rect = nifBoundingRect(swr.nif1Element, refPosition)
        const nif2Rect = nifBoundingRect(swr.nif2Element, refPosition)
        // Determine the "length" of this particular wire and where it starts
        // and ends. For the following calculations we assume that the start is
        // always before then end, that is, the start is located above the end,
        // not below for a vertical swim-lane layout. However, some wires have a
        // natural orientation, such as MACVLAN going from master to MACVLAN
        // network interface. So, when normalizing the start and end we need to
        // remember in order to later to be able to correctly draw according to
        // the original orientation.
        swr.reverse = nif1Rect.top > nif2Rect.top
        const topWire = swr.reverse ? nif2Rect : nif1Rect
        const bottomWire = swr.reverse ? nif1Rect : nif2Rect
        swr.from = topWire.top + (topWire.bottom - topWire.top) / 2
        swr.to = bottomWire.top + (bottomWire.bottom - bottomWire.top) / 2
        swr.len = Math.round(swr.to - swr.from) // later simplifies sorting in case of a tie in length...
        // Also determine the horizontal inset of the network interface badges.
        const fromInset = refWidth - nif1Rect.right
        const toInset = refWidth - nif2Rect.right
        if (!swr.reverse) {
            swr.fromInset = fromInset
            swr.toInset = toInset
        } else {
            swr.fromInset = toInset
            swr.toInset = fromInset
        }
        swr.tooltip = '' // TODO:
        return swr
    })
        // Remove any wires that we couldn't layout.
        .filter(w => w)
        // Sort the wire according to their length and then wires of the same
        // length according to their position.
        .sort((w1, w2) => (w1.len - w2.len) || (w1.from - w2.from))
    // Now let's play ... FROGGER! Well, kind of.
    wiresInLanes.forEach(wire => {
        for (let lane = 0; ; lane++) {
            const wiresInThisLane = lanes[lane]
            if (wiresInThisLane) {
                if (!wiresInThisLane.find(placedWire => (wire.to > placedWire.from) || (wire.from < placedWire.to))) {
                    // no overlap, we can use this lane.
                    wire.lane = lane
                    lanes[lane].push(wire)
                    return // proceed with placing the next wire...
                }
                // continue searching in the next lane for some room to place
                // our frogger in...
            } else {
                // we've found no suitable place in the existing lanes, so add a
                // new lane and place just our wire into it for the moment.
                wire.lane = lane
                lanes[lane] = [wire]
                return // proceed with placing new wire...
            }
        }
    })
    // Done, all wires have been placed in some lane. Someone else might now
    // create SVG elements representing the wires using our layout information.
    return [wiresInLanes, lanes.length]
}

/**
 * Layouted "external" wire which goes from a network interface to the
 * "outside".
 */
interface ExternalWire {
    /** DOM element representing the externally connected network interface. */
    nifElement: HTMLElement
    /** 
     * kind (type) of network interfaces, such as "veth", "macvlan", or "" for
     * wires going to the outside.
     */
    kind: string
    /** upper Y position of wire, where it starts. */
    from: number
    /** upper X inset of wire start. */
    fromInset: number
    /** TODO: tooltip to associate with this wire. */
    tooltip: string
    /** outgoing network interface is down. */
    operStateDown: boolean
}

/**
 * Layout the external facing wires.
 * 
 * @param wires external wires high-level description.
 * @param nifContainerDomId DOM element identifier of container with network
 * interface badges.
 *
 * @returns layouted external facing wires.
 */
const layoutExternals = (wires: Wire[], nifContainerDomId: string): ExternalWire[] => {
    const refPosition = document.getElementById(nifContainerDomId).getBoundingClientRect()
    const refWidth = refPosition.right - refPosition.left
    return wires.map(w => {
        const wr = {
            nifElement: document.getElementById(w.nif1Id),
            kind: w.kind,
            operStateDown: w.operStateDown,
        } as ExternalWire
        if (!wr.nifElement) {
            return null
        }
        // Get the bounding boxes of the network interface badges, relative to
        // some outer DOM container.
        const nifRect = nifBoundingRect(wr.nifElement, refPosition)
        wr.from = (nifRect.top + nifRect.bottom) / 2
        wr.fromInset = refWidth - nifRect.right
        wr.tooltip = '' // TODO:
        return wr
    }).filter(w => w)
}

// Ensures that the "hot" wire SVG elements gets sorted to the end in order to
// be drawn on top of earlier SVG elements.
const sortHotJSXElements = (els: JSX.Element[], hotWires: string[], domIdBase: string) => (
    // We need a stable sort and we want to sort hot element near the end, but
    // keep the order stable inbetween hot elements. So we first need to gather
    // the needed additional properties and only then sort.
    els.map((el, idx) => [
        idx,
        hotWires.includes((el.props.className as string).split(" ")
            .find((cls) => isRelationClassName(domIdBase, cls))),
        el,
    ])
        .sort(([idxA, hotA, elA]: [number, boolean, JSX.Element], [idxB, hotB, elB]: [number, boolean, JSX.Element]) => {
            if (hotA !== hotB) return hotA ? 1 : -1
            return idxA - idxB
        })
        .map(([idx, hot, el]) => el)
)

// The "namespace" used for <marker> SVG elements inside the current DOM
// element ID context, to separate them from other element IDs.
const markerSpace = 'marker-'

/**
 * Renders the SVG path of a (virtual) wire connecting two network interfaces.
 *
 * @param slw layout information for wire.
 * @param laneWidth width of a single lane.
 */
const TwoEndedWire = (slw: SwimlaneWire, laneWidth: number, domIdBase: string, wireClass: string, ghostwireClass: string) => {
    const width = (slw.lane + 0.5) * laneWidth
    const radius = laneWidth / 2
    let path = ''
    switch (slw.kind) {
        case 'macvlan':
        case 'vxlan':
            if (slw.reverse) {
                // The master interface is located above the MACVLAN slave
                // interface on the breadboard. So, the "bump in the wire" is at
                // the top where the master interface is.
                path = `
M ${-slw.fromInset + radius} ${slw.from}
a ${radius} ${radius} 0 0 0 ${radius} ${radius}
l ${slw.fromInset + width - 3 * radius} 0
a ${radius} ${radius} 0 0 1 ${radius} ${radius}
l 0 ${slw.len - laneWidth - radius}
a ${radius} ${radius} 0 0 1 ${-radius} ${radius}
l ${-(slw.toInset + width - radius)} 0`
            } else {
                // The master interface is located below the MACVLAN slave
                // interface on the breadboard. So, the "bump in the wire" is at
                // the bottom where the master interface is.
                path = `
M ${-slw.fromInset} ${slw.from}
l ${slw.fromInset + width - radius} 0
a ${radius} ${radius} 0 0 1 ${radius} ${radius}
l 0 ${slw.len - laneWidth - radius}
a ${radius} ${radius} 0 0 1 ${-radius} ${radius}
l ${-(slw.toInset + width - 3 * radius)} 0
a ${radius} ${radius} 0 0 0 ${-radius} ${radius}`
            }
            break
        default:
            // Default is a wire with rounded corners.
            path = `
M ${-slw.fromInset} ${slw.from}
l ${slw.fromInset + width - radius} 0
a ${radius} ${radius} 0 0 1 ${radius} ${radius}
l 0 ${slw.len - laneWidth}
a ${radius} ${radius} 0 0 1 ${-radius} ${radius}
l ${-(slw.toInset + width - radius)} 0`
    }
    const down = slw.operStateDown ? 'down' : ''

    const relation = relationClassNameFromIds(domIdBase, slw.nif1Element.id, slw.nif2Element.id)
    const marker = `url(#${domIdBase}${markerSpace}${slw.kind})`

    return [
        <path
            key={`ghostwire-${slw.nif1Element.id}-${slw.nif2Element.id}`}
            className={clsx(ghostwireClass, relation)}
            d={path}
        />,
        <path
            key={`${slw.nif1Element.id}-${slw.nif2Element.id}`}
            className={clsx(wireClass, slw.kind, relation, down)}
            markerStart={slw.reverse ? marker : undefined}
            markerEnd={slw.reverse ? undefined : marker}
            d={path}
        />
    ]
}

/**
 * Returns the SVG elements for rendering an externally-facing wire.
 *
 * @param ew data object describing an externally-facing wire.
 * @param lanes the number of lanes having been allocated; this tells us how
 * wide the wiring gutter is, in order to make the externally-facing wires
 * properly horizontally stretch beyond the gutter. 
 * @param laneWidth width of a single "lane".
 * @param domIdBase DOM element ID base string (=prefix), required to correctly
 * reference the correct marker ID.
 * @param wireClass one or more CSS class names to assign to the wire SVG
 * element.
 */
const ExternallyFacingWire = (
    ew: ExternalWire, lanes: number, laneWidth: number,
    domIdBase: string,
    wireClass: string, ghostwireClass: string) => {

    const width = (lanes + 0.5) * laneWidth
    const path = `M ${-ew.fromInset} ${ew.from} l ${width + ew.fromInset} 0`
    const down = ew.operStateDown ? 'down' : ''

    const relation = relationClassNameFromIds(domIdBase, ew.nifElement.id)

    return [
        <path
            key={`ghostwire-${ew.nifElement.id}-ext`}
            className={clsx(ghostwireClass, relation)}
            d={path}
        />,
        <path
            key={`${ew.nifElement.id}-ext`}
            className={clsx(wireClass, 'external', relation, down)}
            markerEnd={`url(#${domIdBase}${markerSpace}external${down})`}
            d={path}
        />
    ]
}

const WireArea = styled('svg')(({ theme }) => ({
    // A "ghost wire" (sic!) in this context is a transparent wire set
    // immediately below the "real virtual" wire, the latter being visible to
    // the user. The reason for an additional transparent wire is that it
    // receives stroke pointer events even if being invisible. In contrast, a
    // dashed wire would receive pointer events only for the dashes, but not in
    // between the dashes. So we need more transparency. Who doesn't, except for
    // shady scrum?
    '& .ghostwire': {
        fill: 'none',
        // a stroke of "transparent" will still receive "stroke" pointer events,
        // as opposed to using a "none" stroke. Now I finally see reason for the
        // differentiation between "none" and "transparent", it didn't took a
        // whole lifetime. Phew.
        stroke: 'transparent !important',
        strokeDasharray: 'none !important',
        strokeDashoffset: '0 !important',
        strokeWidth: '5px',
    },

    '& .wire': {
        // And now for the "real" wires, the ones that are going to be visible to
        // users. These always lay on top of their ghost pendants.
        pointerEvents: 'none',
        fill: 'none',
        stroke: '#000', // fallback color, maybe pink would be better :p
        strokeWidth: '5px',
        strokeLinecap: 'butt',
    },
    '& .wire.down:not(external)': {
        stroke: `${theme.palette.wire.down}`,
    },
    '& .wire.external': {
        stroke: theme.palette.wire.external,
        strokeDasharray: '1.5ex 1ex',
    },
    '& .wire.external.down': {
        strokeOpacity: 1,
    },
    '& .wire.pfvf': {
        stroke: theme.palette.wire.pfvf,
        strokeDasharray: '1.5ex 0.5ex 0.5ex 0.5ex',
    },
    '& .wire.veth': {
        stroke: theme.palette.wire.veth,
    },
    '& .wire.macvlan': {
        stroke: theme.palette.wire.maclvan,
        strokeDasharray: '2.5ex 0.5ex 1ex 0.5ex',
    },
    '& .wire.vxlan': {
        stroke: theme.palette.wire.vxlan,
        strokeDasharray: '1ex 0.5ex',
    },

    '& .ext-marker': {
        fill: theme.palette.wire.external,
    },
    '& .ext-marker-down': {
        fill: `${theme.palette.wire.down} !important`,
    },

    '& .macvlan-marker': {
        fill: theme.palette.wire.maclvan,
    },
    '& .macvlan-marker-down': {
        fill: `${theme.palette.text.disabled} !important`,
    },
}))


/**
 * Takes a marker SVG element and creates two copies of it, one for the marker
 * as-is, another one for the marker variant in operstate "down". The returned
 * markers have their CSS class set to the specified marker class name (for all
 * operstates execpt "down") and for the operstate down, respectively.
 *
 * @param id unique DOM element id to be assigned to the returned marker
 * elements, once exactly as is, and the second marker element for the operstate
 * "down" version of the marker.
 * @param cname CSS class name for marker in operstates other than "down".
 * @param cnamedown CSS class name for marker in operstate "down".
 * @param m marker SVG element acting as template.
 */
const marker = (id: string, cname: string, cnamedown: string, m: ReactElement) =>
    [
        React.cloneElement(m, { id: id, className: cname, key: id }),
        React.cloneElement(m, { id: `${id}down`, className: cnamedown, key: `${id}down` }),
    ]

const externalMarker = <marker
    stroke="none"
    markerWidth="3" markerHeight="3"
    refX="0" refY="1.5">
    <path d="M 0 3 L 3 1.5 L 0 0 Z" />
</marker>

const macvlanMarker = <marker
    stroke="none"
    markerWidth="2" markerHeight="2"
    refX="1" refY="1">
    <path d="M 1 2 a 1 1 0 0 0 0 -2 a 1 1 0 0 0 0 2" /> {/* correct, that's a circle */}
</marker>


export interface WiringProps {
    /** wires to render between network interfaces. */
    wires: Wire[]
    /** class name(s) of "hot" wire(s), that is, selected wire(s) */
    hotWires: string[]
    /** CSS class(es). */
    className?: string
    /** 
     * changing the "layout token" triggers (re-) layouting the wires, even if
     * the wire configuration hasn't changed, yet the placement of the network
     * interface badge DOM elements might have changed. Any number here
     * suffices, as long as it changes whenever a re-layout is needed.
     */
    layoutToken: any
}

/**
 * `Wiring` renders a wire "pane" as well as the wiring on top of it. The
 * wiring is described using `Wire` objects; bascially, the wire objects
 * describe two (one) network interface DOM element to wire with each other,
 * as well as the "kind" of wire (such as "veth", "macvlan", et cetera.)
 *
 * The wire pane will automatically resize its width in order to accommodate
 * the wires, this is determined when the component layouts the wires.
 *
 * The `Wiring` needs to be triggered for re-layouting the wires by changing
 * its `layoutToken` property whenever the (separate) "content" component
 * rendering the breadboard content including the network interfaces changes
 * in its dimensions. Thus, the `Breadboard` component automatically takes
 * care of correctly triggering its `Wiring` child component.
 *
 * Please note that the wires on the wire pane will overflow to the left of
 * the pane in order to properly connect with the outlines of the network
 * interface DOM elements.
 */
export const Wiring = ({ wires, hotWires, className, layoutToken }: WiringProps) => {

    const nifContainerDomId = useContextualId('breadboard-content')
    const domIdBase = useContextualId('')

    const laneWidth = parseInt(useTheme().spacing(2))

    const [lanesCount, setLanesCount] = useState(1)
    const [svgWires, setSvgWires] = useState([])

    // Separate outgoing wires from point-to-point wires: the latter need
    // swim-lane layouting, while the former are just going off the grid...
    const externalWires = wires.filter(w => !w.nif2Id)
    const p2pWires = wires.filter(w => w.nif2Id)

    // When the wiring information changes, we need to rerender the SVG wiring
    // components. For this, we need to determine where the network interfaces
    // have been rendered for display.
    useLayoutEffect(() => {
        const [p2pw, lc] = layoutWires(p2pWires, nifContainerDomId)
        const extw = layoutExternals(externalWires, nifContainerDomId)
        setSvgWires(
            sortHotJSXElements(
                extw.map(ew =>
                    ExternallyFacingWire(ew, lc, laneWidth, domIdBase, 'wire', 'ghostwire')).flat()
                    .concat(p2pw.map(w =>
                        TwoEndedWire(w, laneWidth, domIdBase, 'wire', 'ghostwire'))
                        .flat()),
                hotWires, domIdBase)
        )
        // Add room for one extra lane room for external facing wires.
        setLanesCount(lc + 2)
        // Most of the following dependencies most probably don't make sense
        // in a practical sense of things that really are going to change as
        // opposed there is theoretical chance of things going to change and
        // so we have to include them, because everyone knows it is best
        // practise that things we know to not change will change and break
        // ... except that in real life even things don't break all the time.
        // How often do we expect classes and classes.wire to change?!!
        // Seriously!

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [layoutToken,
        // These are breaking our necks: p2pWires, externalWires, 
        hotWires,
        laneWidth, nifContainerDomId, domIdBase])

    // Render the rendered SVG wires. This additionally needs (start/end)
    // marker definitions to be also present. Now, SVG 1.x has a major design
    // deficit when it comes to styling markers using CSS in that markers need
    // to be referenced by element identifier, so the classes applied to the
    // path referencing the marker element do not carry over; one of several
    // SVG SNAFUs. We have to work around this by creating multiple versions
    // of the same marker, one for down operstate, and another one for the
    // other operstates.
    return (
        <WireArea className={className} width={lanesCount * laneWidth}>
            <defs>
                {marker(
                    `${domIdBase}${markerSpace}external`,
                    'ext-marker', 'ext-marker-down',
                    externalMarker
                )}
                {marker(
                    `${domIdBase}${markerSpace}macvlan`,
                    'macvlan-marker', 'macvlan-marker-down',
                    macvlanMarker
                )}
            </defs>
            <g>
                {svgWires}
            </g>
        </WireArea>
    )
}

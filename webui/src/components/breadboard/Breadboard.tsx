// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useState, useRef, useMemo } from 'react'

import { darken, lighten, styled } from '@mui/material'
import { keyframes } from '@mui/system'
import useResizeObserver from 'beautiful-react-hooks/useResizeObserver'

import { useContextualId } from 'components/idcontext'
import { Wire, Wiring } from 'components/wiring'
import { NetworkInterface, NetworkNamespace, NetworkNamespaces, nifId, OperationalState } from 'models/gw'
import { isRelationClassName } from 'utils/relclassname'
import { rgba } from 'utils/rgba'


// The dash array to use for drawing "hot wire" paths.
const hotwireDashArr = [16, 8]
const hotwireTotalDashlen = hotwireDashArr.reduce((sum, val) => sum + val, 0)


// The guiding principle of our marching ants animation is "one step
// forward, one step back". In between steps we put a small pause. And we
// additionally ease in and out the step animation, but with a longer
// intermediate phase compared to what the build-in CSS ease timing function
// does.
const marchingAnts = keyframes({
    '0%': {
        strokeDashoffset: 0,
        animationTimingFunction: 'cubic-bezier(.4,0,.8,.5)',
    },
    '16%': {
        strokeDashoffset: -1 * hotwireTotalDashlen,
        animationTimingFunction: 'linear',
    },
    '23%': {
        strokeDashoffset: -2 * hotwireTotalDashlen,
        animationTimingFunction: 'cubic-bezier(.2,.5,.6,1)',
    },
    '39%': {
        strokeDashoffset: -3 * hotwireTotalDashlen,
        animationTimingFunction: 'linear',
    },
    '50%': {
        strokeDashoffset: -3 * hotwireTotalDashlen,
        animationTimingFunction: 'cubic-bezier(.4,0,.8,.5)',
    },
    '66%': {
        strokeDashoffset: -2 * hotwireTotalDashlen,
        animationTimingFunction: 'linear',
    },
    '73%': {
        strokeDashoffset: -1 * hotwireTotalDashlen,
        animationTimingFunction: 'cubic-bezier(.2,.5,.6,1)',
    },
    '89%': {
        strokeDashoffset: 0,
        animationTimingFunction: 'linear',
    },
    '100%': {
        strokeDashoffset: 0,
    }
})

// There isn't much styling going on with respect to the Breadboard component
// itself, because it basically is simply a "content pane" (placed on the left
// side for ltr, but we don't support rtl) which then is accompanied by a
// "wiring pane" sitting right (sic!) next to it.
//
// On second thoughts ... never mind.
//
// Traditionally, electronics experiments with and without using breadboards
// were always build upon kitchen tables...
const KitchenTable = styled('div')(({ theme }) => ({
    display: 'flex',
    flexDirection: 'row',

    // And now for some CSS specifity madness: while it is tempting to
    // simply throw in the nuclear "!important" option, this conflicts with
    // the ghost wires (as opposed to the non-ghost, visible wires). We must
    // not change these ghost wires. By raising the specifity of this rule
    // we make sure that we can override the styling of ordinary wires, yet
    // not the ghost wires (which are thus using "!important" to protect
    // their own styling).
    '& path.hot:not(.somethingcompletelydifferent):not(.else)': {
        stroke: theme.palette.wire.hot,
        strokeDasharray: hotwireDashArr.join(" "), // ...aaarrrrrrgh, that's why that package is called "emotion"
        strokeOpacity: 1,
        strokeWidth: '6px',
        animation: `${marchingAnts} 4s infinite`,
    },
    '& .hot > span:not(.nifbagdeicon)': {
        borderColor: theme.palette.wire.hot,
        backgroundColor: rgba(theme.palette.wire.hot, .05),
    },
    '& .hot > button.MuiButtonBase-root': {
        borderColor: theme.palette.wire.hot,
        backgroundColor: rgba(theme.palette.wire.hot, .05),
    },

    '& path': {
        transition: 'stroke 0.4s ease-in-out',
    },
    '& g.hot > path:not(.hot)': {
        stroke: theme.palette.mode === 'dark'
            ? lighten(theme.palette.background.default, 0.15)
            : darken(theme.palette.background.default, 0.1),
    },
}))

// Allow the content pane to grow as necessary, as to snatch up any free
// horizontal room.
const ContentPane = styled('div')(({ theme }) => ({
    flexGrow: 1,
}))

// The wiring pane will automatically size itself horizontally to the
// width needed in order to accommodate the wiring. This can be done only
// programmatically, as SVG doesn't happen to do "auto" size calculations.
const WiringPane = styled(Wiring)(({ theme }) => ({
    overflow: 'visible',
    flexGrow: 0,
    flexShrink: 0,
    zIndex: 10,
}))

export interface BreadboardProps {
    /** 
     * network namespaces for which to generate, layout and show the wiring.
     * The `Breadboard` component accepts not only the usualy discovery
     * network namespaces map, but also an array of network namespaces or even
     * a single network namespace.
     */
    netns: NetworkNamespaces | NetworkNamespace[] | NetworkNamespace
    /** children to render within the content pane. */
    children: React.ReactNode
}

/**
 * "Extract" the wiring information from a single network namespace or (more
 * typical) from a set of network namespaces, thus from the discovered virtual
 * networking topology.
 *
 * @param netns network namespaces for which to return the wires information.
 * This can be the usualy discovery network namespaces map, but also an array of
 * network namespaces or even a single network namespace.
 */
const extractWiring = (
    netns: NetworkNamespaces | NetworkNamespace[] | NetworkNamespace,
    domIdBase: string
) => {
    // To start with, bring the specified network namespace(s) into our
    // canonical form of an array of network namespaces.
    var netnses: NetworkNamespace[]
    if (Array.isArray(netns)) {
        netnses = netns
    } else if ('netnsid' in netns) {
        netnses = [netns]
    } else {
        netnses = Object.values(netns)
    }
    // Now skim through all network namespaces with their network interfaces
    // and generate wire information. In order to not create two-ended wires
    // twice, we need to remember (some of) the wires we already created...
    const wires = new Map<NetworkInterface, Wire>()
    netnses.forEach(netns =>
        // Please note that "lo" loopbacks never get an external facing wire ...
        // this would look really awkward.
        Object.values(netns.nifs).filter(nif => nif.name !== 'lo').forEach(nif => {
            // For (virtual) hardware network interface, we create a wire
            // going off into reality ... or whatever we get to see of it.
            if (nif.isPhysical) {
                // If this is a VF, then we instead add its relationship to its
                // PF...
                if (nif.pf) {
                    wires.set(nif, {
                        kind: "pfvf",
                        operStateDown: (nif.operstate === OperationalState.Down) || (nif.pf.operstate === OperationalState.Down),
                        nif1Id: domIdBase + nifId(nif),
                        nif2Id: domIdBase + nifId(nif.pf),
                    } as Wire)
                } else {
                    // It's either a non-SR-IOV network interface or a PF: show
                    // an outward going line.
                    wires.set(nif, {
                        kind: '',
                        operStateDown: nif.operstate === OperationalState.Down,
                        nif1Id: domIdBase + nifId(nif),
                    } as Wire)
                }
            } else switch (nif.kind) {
                case 'veth':
                    if (nif.peer && !wires.has(nif.peer) && !wires.has(nif)) {
                        wires.set(nif, {
                            kind: nif.kind,
                            operStateDown: nif.operstate === OperationalState.Down,
                            nif1Id: domIdBase + nifId(nif),
                            nif2Id: domIdBase + nifId(nif.peer),
                        } as Wire)
                    }
                    break;
                case 'macvlan':
                    wires.set(nif, {
                        kind: nif.kind,
                        operStateDown: (nif.operstate === OperationalState.Down) || (nif.macvlan.operstate === OperationalState.Down),
                        nif1Id: domIdBase + nifId(nif),
                        nif2Id: domIdBase + nifId(nif.macvlan),
                    } as Wire)
                    break;
                case 'vlan':
                    wires.set(nif, {
                        kind: nif.kind,
                        operStateDown: nif.operstate === OperationalState.Down,
                        nif1Id: domIdBase + nifId(nif),
                        nif2Id: domIdBase + nifId(nif.master),
                    } as Wire)
                    break;
                case 'vxlan':
                    wires.set(nif, {
                        kind: nif.kind,
                        operStateDown: nif.operstate === OperationalState.Down,
                        nif1Id: domIdBase + nifId(nif),
                        nif2Id: domIdBase + nifId(nif.underlay),
                    } as Wire)
                    break;
                // TODO: TAPTUN, ...
            }
        }))
    return [...wires.values()]
}

/**
 * Returns the relation CSS class name(s) of an element or one of its parents
 * (only within a limited reach of four parents), if any are found. Otherwise,
 * returns null. The search for relation CSS class name(s) on parents of a
 * target element is necessary since mouseover (+...) events do not bubble, so
 * we might see them on some child of a network interface badge instead.
 *
 * @param el a DOM element, usually an event target object.
 * @param domIdBase DOM element ID context (namespace, hehe).
 */
const locateTargetRelationClasses = (el: Element, domIdBase: string) => {
    for (var hierarchy = 1; hierarchy <= 5 && el; hierarchy++) {
        const classNames = [...el.classList]
        const relations = classNames
            .filter(className => isRelationClassName(domIdBase, className))
        if (relations.length) {
            return relations
        }
        el = el.parentElement
    }
    return []
}

/**
 * The `Breadboard` component renders its children components and then wires up
 * the network interfaces in all those rendered children, according to a wiring
 * plan derived from the network namespaces passed to this component via its
 * `netns` property.
 *
 * This component additionally supports highlighting individual wires by setting
 * the "hot" CSS class on the DOM elements for the wires. When marking DOM (SVG)
 * path elements as "hot" they show marching ants, going rhythmically forth and
 * back.
 */
export const Breadboard = ({ children, netns }: BreadboardProps) => {

    const domIdBase = useContextualId('')
    const nifContainerDomId = domIdBase + 'breadboard-content'

    // List of class names of DOM wire elements that are currently highlighted,
    // if any.
    const [hotWireClasses, setHotWireClasses] = useState<string[]>([])
    const [selected, setSelected] = useState(false)

    const breadboardref = useRef()

    const wires = extractWiring(netns, domIdBase)

    // beautiful react hooks to the (layouting) rescue, as they offer us a slick
    // resize observer with integrated debouncing/throttling. We need to
    // relayout the wires not only when the wiring changes, but also when
    // network interfaces move around, such as when the user resizes the
    // breadboard. In order to trigger layouting and subsequently rerendering
    // the wires in those circumstances, we simply derive a "layout token"
    // changing whenever the dimensions of the content change (as this then can
    // shuffle the wired network interfaces around) and we need to follow.
    const contentref = useRef()
    const contentRect = useResizeObserver(contentref, 100/*ms*/)
    const layoutToken = contentRect ? `${contentRect.width}x${contentRect.height}` : 0

    const contentMemo = useMemo(() => (
        <ContentPane id={nifContainerDomId} ref={contentref}>
            {children}
        </ContentPane>
    ), [nifContainerDomId, children])

    // tag or untag the wires identified by their wire classes as "hot" (to
    // highlight) or "cold" (normal appearance).
    const tagHotElements = (wireClasses: string[], hot: boolean) => {
        const els = wireClasses
            .map(className => [...(breadboardref.current as HTMLDivElement).getElementsByClassName(className)])
            .flat()
        if (hot) {
            els.forEach(el => {
                el.classList.add('hot')
                if (el instanceof SVGElement) {
                    el.parentElement.classList.add('hot')
                }
            })
            setHotWireClasses(wireClasses)
        } else {
            els.forEach(el => {
                el.classList.remove('hot')
                if (el instanceof SVGElement) {
                    el.parentElement.classList.remove('hot')
                }
            })
            setHotWireClasses([])
        }
    }

    // highlights a set of wired network interfaces with their "wire" relation
    // or removes the highlighting from them. The related network interfaces
    // are glanced from specially formed CSS class names attached to network
    // interface and wire DOM elements.
    const hotOrCold = (e: React.MouseEvent, hot?: boolean) => {
        if (selected || !breadboardref.current) {
            return
        }
        const hotWireClasses = locateTargetRelationClasses(e.target as Element, domIdBase)
        if (!hotWireClasses.length) {
            return
        }
        tagHotElements(hotWireClasses, hot)
    }

    const handleMouseOver = (e: React.MouseEvent) => {
        hotOrCold(e, true)
    }

    const handleMouseOut = (e: React.MouseEvent) => {
        hotOrCold(e)
    }

    const handleClick = (e: React.MouseEvent) => {
        if (e.target instanceof SVGElement) {
            const clickedWireClasses = locateTargetRelationClasses(e.target as Element, domIdBase)
            if (clickedWireClasses.length) {
                // User clicked on a wire: if this is the currently selected
                // "hot" wire then deselect it again. If it's a different wire,
                // then deselect the current hot wire and select the newly
                // clicked wire.
                tagHotElements(hotWireClasses, false) // deselect any old wire
                if (!selected
                    || clickedWireClasses.length !== hotWireClasses.length
                    || clickedWireClasses.find((el, idx) => el !== hotWireClasses[idx])) {
                    // it's a different selection, so select the new wire.
                    setSelected(true)
                    tagHotElements(clickedWireClasses, true)
                    return
                }
            }
        }
        // Clicked somewhere else, so unselect and enable mouse overs again.
        tagHotElements(hotWireClasses, false)
        setSelected(false)
    }

    return (
        <KitchenTable
            ref={breadboardref}
            onMouseOver={handleMouseOver}
            onMouseOut={handleMouseOut}
            onClick={handleClick}
        >
            {contentMemo}
            <WiringPane
                wires={wires}
                hotWires={hotWireClasses}
                layoutToken={layoutToken} />
        </KitchenTable>
    )
}

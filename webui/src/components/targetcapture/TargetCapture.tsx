// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { Button, styled, Theme } from '@mui/material'

import { NetworkInterface, NetworkNamespace, Containee, isPod, isNetworkNamespace, isNetworkInterface, isContainee, isOperational, firstContainee, containeeType, PrimitiveContainee, Pod, isContainer, Container } from 'models/gw'
import { CaptureIcon } from 'icons/Capture'
import { useDynVars } from 'components/dynvars'
import { rgba } from 'utils/rgba'
import { Target } from 'models/packetflix/target'
import { keyframes } from '@mui/system';


// During development The (Capturing) Monolith can be enabled using
// REACT_APP_ENABLE_MONOLITH.
const forceMonolith = !!import.meta.env.REACT_APP_ENABLE_MONOLITH

// Calculate the static part of any remote live packet capture URL; it bases on
// the base URI of the application, so we only calculate it once when this
// module gets loaded.
const baseURI = new URL(document.baseURI)
switch (baseURI.protocol) {
    case 'http:':
        baseURI.protocol = 'ws:'
        break
    case 'https:':
        baseURI.protocol = 'wss:'
}
baseURI.pathname += baseURI.pathname.endsWith("/") ? "" : "/"


const bubble = (theme: Theme) => rgba(theme.palette.capture, theme.palette.mode === 'light' ? 0.4 : 0.8)

const bubbleKeyframes = keyframes({ // bubbles moving up
    '0%': { transform: 'translate()' },
    '100%': { transform: 'translate(0, -80%)' },
})

const floodKeyframes = keyframes({ // water level raising from bottom to half.
    '0%': { transform: 'translate(0, 50%)' },
    '100%': { transform: 'translate()' },
})

const swayKeyframes = keyframes({ // waves swaying forth and back...
    '0%': { transform: 'translate()' },
    '50%': { transform: 'translate(-10px, 0)' },
    '100%': { transform: 'translate()' },
})

const SharkButton = styled(Button)(({ theme }) => ({
    // make the button circular: fix minimal width, as otherwise the
    // icon will be boxed to the left and right; set a much-too-large border
    // radius, so it automatically gets set to the correct radius based on
    // the element's actual size.
    borderRadius: '50px',
    borderColor: 'transparent !important', //theme.palette.capture,
    color: theme.palette.capture,
    // Button (icon) padding needs to be adjusted based on the button's size
    // style; we need to override the default padding as otherwise we get
    // increased left and right padding, turning the button into an oblong
    // shape -- which we don't want (we're not following the way of the
    // Oblong, y'know).
    '&.smallsize': {
        minWidth: '20px',
        padding: '3px 3px',
    },
    '&.mediumsize': {
        minWidth: '26px',
        padding: '5px 5px',
    },
    // Ensure that the bubbles get clipped to the circle outline of our
    // capture button.
    overflow: 'hidden',
    // As for the order of and colons required for "hover" (=pseudo class)
    // and "before" (pseudo element), see also:
    // https://stackoverflow.com/a/5777334
    '&:hover::before': {
        content: '""',
        pointerEvents: 'none',
        opacity: 0.75,
        background: `
                radial-gradient(circle at 10% 50%, transparent 0, transparent 1px, ${bubble(theme)} 2px, ${bubble(theme)} 3px, transparent 3px),
                radial-gradient(circle at 55% 40%, transparent 0, transparent 2px, ${bubble(theme)} 3px, ${bubble(theme)} 4px, transparent 4px),
                radial-gradient(circle at 80% 70%, transparent 0, transparent 1px, ${bubble(theme)} 2px, ${bubble(theme)} 3px, transparent 3px)
            `,
        width: '100%',
        height: '300%',
        top: 0,
        left: 0,
        position: 'absolute',
        animation: `${bubbleKeyframes} 4s linear infinite both`,
    },
    '&:hover::after': {
        content: '""',
        pointerEvents: 'none',
        background: `
                radial-gradient(circle at 4px -3px, transparent 4px, ${rgba(theme.palette.primary.light, 0.3)} 5px)
            `,
        backgroundSize: '8px 12px',
        backgroundRepeat: 'repeat-x',
        width: '200%',
        height: '100%',
        top: '55%',
        left: '-10%',
        position: 'absolute',
        animation: `${floodKeyframes} 0.5s ease 1 forwards, ${swayKeyframes} 2s ease infinite`,
    },
}))


export interface TargetCaptureProps {
    /** capture target */
    target: Containee | NetworkInterface | NetworkNamespace
    /** button size */
    size?: 'small' | 'medium'
    /** CSS class name(s) */
    className?: string
    /** for use in help, always shows up, doesn't start capture. */
    demo?: boolean
}

/**
 * Renders a "packetflix" capture button, which when clicked triggers a
 * hand-over to a client system-installed external capture plugin for Wireshark
 * to start a live packet capture session (using a "packetflix:" scheme URL).
 *
 * There are different ways to select capture targets, that is, a network
 * namespace and optionally one or all of its network namespaces:
 *
 * 1. containee ... and thus capturing from all the network interfaces of the
 *    containee's network namespace.
 * 2. individual network interface ... which always belongs to exactly one
 *    network namespace.
 * 3. network namespace ... with all of its (currently present) network
 *    interfaces.
 *
 * While specifying a containee or network namespace are mostly equal in most
 * situations, there are subtle differences in how capture target is eventually
 * (re)identified in case the system state changes between discovery and start
 * of the capture: in particular, when the specified network namespace doesn't
 * exist anymore or might get reused by a different containee.
 *
 * - containee: in case the containee is a process or container and it has
 *   changed since the discovery, then the capture service will transparently
 *   rerun a simplified scan to find the current network namespace attached to
 *   the new containee with the given name.
 *
 * - network interface: in case the network namespace and containee changes to
 *   which the network interface belongs, then the capture service will
 *   transparently rerun a simplified scan to find the current network namespace
 *   attached to the new containee with the given name.
 *
 * - network namespace: the lexicographically earliest containee or init(1) is
 *   selected (depending on situation) as the leading reference; form there on
 *   consistency checks and re-lookup will work the same as when specifying a
 *   containee directly.
 */
export const TargetCapture = ({ target, size = 'medium', className, demo }: TargetCaptureProps) => {
    // Is capturing enabled in the UI?
    const { enableMonolith } = useDynVars()

    // Render nothing unless capture links have been enabled via a global
    // variable set for this application.
    if (!(enableMonolith || forceMonolith || demo) || !target) {
        return null
    }

    const netns: NetworkNamespace | undefined =
        (isNetworkNamespace(target) && target)
        || (isNetworkInterface(target) && target.netns)
        || (isContainee(target) && (isPod(target) ? target.containers[0].netns : target.netns))
        || undefined
    if (!netns && !demo) {
        return null
    }

    // We never hide the capture button, but we'll disable it for a network
    // interface target which isn't operational.
    const biting = !isNetworkInterface(target) || isOperational(target)

    // Unless a specific network interface is targetted, capture from all
    // "working" network interfaces of the network namespace involved.
    const nifs: string[] =
        (isNetworkInterface(target) && [target.name])
        || (!!netns && Object.values(netns.nifs)
            .filter(nif => isOperational(nif))
            .map(nif => nif.name))
        || []

    // Try to get a suitable containee that we can use to later add some useful
    // meta information to the packet capture stream.
    const containee: Containee | undefined =
        (isNetworkNamespace(target) && firstContainee(target))
        || (isNetworkInterface(target) && firstContainee(target.netns))
        || (isContainee(target) && target)
        || undefined

    const targetDetails: Target = {
        netns: netns!.netnsid,
        'network-interfaces': nifs,
        name: containee && containee.name || '',
        type: (!containee && '')
            || (!isPod(containee as Containee) && containeeType(containee as PrimitiveContainee))
            || (isPod(containee as Containee) && (containee as Pod).flavor)
            || '',
        prefix: (!containee && '')
            || (isContainer(containee as Containee) && (containee as Container).turtleNamespace)
            || (isPod(containee as Containee) && (containee as Pod).containers[0].turtleNamespace)
            || '',
    }
    const capturl = `packetflix:${baseURI.toString()}capture`
        + `?container=${encodeURIComponent(JSON.stringify(targetDetails))}`
        + `&nif=${nifs.join("/")}`

    return (
        <SharkButton
            className={clsx(className, `${size}size`)}
            href={demo ? '' : capturl}
            size={size}
            variant="outlined"
            disabled={!biting}
        ><CaptureIcon fontSize="inherit" /></SharkButton>
    )
}

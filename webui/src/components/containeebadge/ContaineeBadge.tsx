// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { Box, Button, styled, Tooltip } from '@mui/material'

import { containeeDescription, containeeState, ContainerState, containerStateString, Containee, containeeFullName, isPrivilegedContainer, isElevatedContainer, isContainer, GHOSTWIRE_LABEL_ROOT } from 'models/gw'
import { ContaineeIcon } from 'utils/containeeicon'
import { TargetCapture } from 'components/targetcapture'
import { PrivilegedIcon } from 'icons/containeestates/Privileged'
import { IEAppDebuggableStateIcon } from 'icons/containeestates/IEAppDebuggableState'
import CapableIcon from 'icons/containeestates/Capable'


// Obviously, there is quite some styling going on here in order to render
// badges for box entities, such containers, stand-alone processes, et cetera.
const Badge = styled('div')(() => ({
    display: 'inline-block',
}))

const CaptureButton = styled(TargetCapture)(({ theme }) => ({
    marginLeft: theme.spacing(0.5),
}))

const CeeBadge = styled(Button)(({ theme }) => ({
    // General+basis badge styling...
    display: 'inline-block',

    paddingLeft: 'calc(0.5em + 8px)', // increase padding to accomodate left thick border.
    paddingRight: '0.5em',
    paddingTop: '0.1ex',
    paddingBottom: '0.1ex',
    borderColor: theme.palette.mode === 'light' ?
        'rgba(0, 0, 0, 0.23)' : 'rgba(255, 255, 255, 0.23)',
    borderWidth: '1px',
    borderLeftWidth: '0',
    borderStyle: 'solid',

    // ...ensure that we DON'T inherit any hanging indentations! :(
    textIndent: '0',

    // line height of Material UI buttons...
    lineHeight: '1.75',

    // Ensures that the ::before pseudo element will show up.
    position: 'relative',

    // To get a left border without a slope where it changes into the
    // top/bottom borders, we need to place the badge's prominent left
    // border block separately before the badge itself; see also:
    // https://stackoverflow.com/a/11052405
    '&::before': {
        // Otherwise, pseudo-elment won't show up.
        display: 'inline-block',
        content: '""',

        // Set a solid, thick left border.
        borderLeftStyle: 'solid',
        borderLeftWidth: '8px',
        borderLeftColor: theme.palette.containee.bindmount,

        // place the pseudo-element's border where the normal border
        // should be; diverging from the SO answer linked to above, we
        // don't want the left border protrude from the badge, so we have
        // to fix the right border here instead and then ensure that
        // there's enough *left* padding inside the badge.
        position: 'absolute',
        top: '-1px',
        bottom: '-1px',
        left: '0px',
        right: '8px',
    },

    // Status colors for the left border indicator, depending on the
    // status of a container or process.
    [`&.${containerStateString(ContainerState.Running)}::before`]: {
        borderLeftColor: theme.palette.containee.running,
    },
    [`&.${containerStateString(ContainerState.Restarted)}::before`]: {
        borderLeftColor: theme.palette.containee.running,
    },
    [`&.${containerStateString(ContainerState.Paused)}::before`]: {
        borderLeftColor: theme.palette.containee.paused,
    },
    [`&.${containerStateString(ContainerState.Exited)}::before`]: {
        borderLeftColor: theme.palette.containee.exited,
    },

    // privileged warning
    '&.privileged': {
        backgroundColor: theme.palette.containee.privileged.main,
        color: `${theme.palette.containee.privileged.contrastText} !important`,
    },
    '& svg.MuiSvgIcon-root:first-of-type': {
        paddingRight: '0.2em',
    },
    '& .MuiSvgIcon-root + .MuiSvgIcon-root': {
        paddingLeft: 0,
        paddingRight: '0.2em',
    },
    '&.elevated .MuiSvgIcon-root.kinginthebox': {
        color: theme.palette.containee.elevated.main,
    },

    '& .MuiSvgIcon-root': {
        verticalAlign: 'middle',
        
    },

    // Ensure to use the proper button appearance when the badge acts as a
    // button, but still subject to optionally enforcing a rectangular
    // shape.
    '&.button': {
        color: theme.palette.text.secondary,
        fontWeight: 'normal',
        textTransform: 'none',
        borderRadius: 0,
        whiteSpace: 'nowrap',
    },
    '&.button:hover': {
        borderLeftWidth: 0,
    },
    '&.button.rounded': {
        borderRadius: theme.shape.borderRadius,
    },
    '&.button.rounded::before': {
        borderTopLeftRadius: theme.shape.borderRadius,
        borderBottomLeftRadius: theme.shape.borderRadius,
    },
}))

export interface ContaineeBadgeProps {
    /** 
     * containee object (bind mount, stand-alone process, container, pod, ...)
     * to render information about.
     */
    containee: Containee
    /** 
     * turns the badge into a clickable button that fires an onclick event
     * with the boxed entity object (container, process) the user has clicked.
     * Unless squared is explicitly set, the badge will show rounded corners
     * following the style language of Material Design for buttons.
     */
    button?: boolean
    /** 
     * optional click handler that will receive the boxed entity object
     * (tenant) which was clicked.
     */
    onClick?: ((containee: Containee) => void)
    /** keeps the angled badge design even with `button` enabled. */
    angled?: boolean
    /** 
     * optional element (usually an icon) to place to the right of the badge
     * contents.
     */
    endIcon?: React.ReactNode
    /** disable tooltip? */
    notooltip?: boolean
    /** show an additional capture button? */
    capture?: boolean
    hideCapture?: boolean
}

/**
 * Component `ContaineeBadge` renders information about a containee, that is, a
 * namespaced entity, such as a stand-alone process or a container. The badge
 * can be either a passive badge or an active button (and thus triggering some
 * action when pressed).
 *
 * Additionally, the containee badge can feature a capture button to start live
 * capture streaming.
 *
 * In case the containee is a privileged container or a container with
 * non-default Docker capabilities, then it will change style to indiciate
 * elevated capabilities.
 *
 * | view use case | button? | end icon? | capture? |
 * | --- | --- | --- | --- |
 * |  wiring | button ☞ containee details | maximize | yes |
 * |  all netns | button ☞ containee details | restore | yes |
 * |  single netns details | - | - | yes |
 * |  related containee | button ☞ containee details | - | - |
 *
 * In the case of an active badge button, this component calls the specified
 * `onClick` callback function, passing the `Containee` instance for this badge.
 * Callback handlers then might choose to (smoothly) scroll the related
 * containee into view, et cetera.
 *
 * - the name of the boxes entity, such as the container or process name.
 * - an icon signalling the type of boxed thing, or the type of container.
 * - the container or process status, in case this is a container "box" or a
 *   stand-alone process. Bind-mounted boxes get the theme's default color
 *   instead.
 * - a tooltip with the name and optional status of the boxed entity (can be
 *   disabled via the `notooltip` property).
 */
export const ContaineeBadge = ({
    containee,
    button,
    angled,
    endIcon,
    notooltip,
    capture,
    hideCapture,
    onClick
}: ContaineeBadgeProps) => {
    const privileged = isPrivilegedContainer(containee)
    const elevated = !privileged && isElevatedContainer(containee)
    const appdebuggable = isContainer(containee) && !!containee.labels[GHOSTWIRE_LABEL_ROOT+'ieapp/debuggable']

    const classNames = clsx(
        containeeState(containee),
        button && 'button',
        button && !angled && 'rounded',
        privileged && 'privileged',
        elevated && 'elevated',
        appdebuggable && 'ieapp-debuggable',
    )

    const CeeIcon = ContaineeIcon(containee)

    const shark = capture && (!hideCapture
        ? <CaptureButton target={containee} />
        : <Box sx={{width: '30px', display:'inline-block'}} />)

    const kingOfBoxIcon = privileged
        ? <PrivilegedIcon className="kinginthebox" />
        : elevated ? <CapableIcon className="kinginthebox" /> : ''

    const bugMeIcon = appdebuggable
        ? <IEAppDebuggableStateIcon className="iebugme" />
        : ''

    // User pressed the badge button, now call the supplied callback function
    // and pass it the containee object.
    const handleClick = () => {
        onClick && onClick(containee)
    }

    const badge = (button &&
        <Badge>
            <CeeBadge
                variant="outlined"
                className={classNames}
                onClick={handleClick}
            >
                <CeeIcon className="typeicon" />{kingOfBoxIcon}{bugMeIcon}{containeeFullName(containee)}{endIcon}
            </CeeBadge>
            {shark}
        </Badge>
    ) || <Badge>
            <CeeBadge as="div" className={classNames}>
                <CeeIcon className="typeicon" />{kingOfBoxIcon}{bugMeIcon}{containeeFullName(containee)}{endIcon}
            </CeeBadge>
            {shark}
        </Badge>

    return notooltip ? badge : <Tooltip title={containeeDescription(containee)}>{badge}</Tooltip>
}

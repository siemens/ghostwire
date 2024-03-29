// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'

import { useAtom } from 'jotai'

import { Button, CircularProgress, Fade, IconButton, Menu, MenuItem, styled, Tooltip } from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh'
import SyncIcon from '@mui/icons-material/Sync'
import SyncDisabledIcon from '@mui/icons-material/SyncDisabled'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'

import { discoveryRefreshingAtom, discoveryRefreshIntervalAtom } from 'components/discovery'
import useId from 'hooks/id'


const defaultThrobberThreshold = 500/* ms */

export interface RefresherInterval {
    /** 
     * interval label to show; if undefined then a suitable label will be
     * derived from the interval number.
     */
    label?: string
    /** interval in milliseconds, or null if off. */
    interval: number | null
}

const defaultIntervals = [
    { interval: null },
    { interval: 500 },
    { interval: 1000 },
    { interval: 5 * 1000 },
    { interval: 10 * 1000 },
    { interval: 30 * 1000 },
    { interval: 60 * 1000 },
    { interval: 5 * 60 * 1000 },
] as RefresherInterval[]

/**
 * Converts an interval number in milliseconds into a suitable textual label,
 * such as 500ms, 30s, et cetera. An interval value of null is taken as
 * "off".
 *
 * @param interval milliseconds or null (=off)
 */
const intervalToLabel = (interval: number | null) => {
    if (interval === null) {
        return "off"
    }
    const ms = interval % 1000
    const t = ms ? [`${ms}ms`] : []
    interval = Math.floor(interval / 1000)
    const sec = interval % 60
    if (sec) {
        t.unshift(`${sec}s`)
    }
    const min = Math.floor(interval / 60)
    if (min) {
        t.unshift(`${min}min`)
    }
    return t.join(' ')
}

const Refreshee = styled('div')(() => ({
    display: 'inline-flex', // keep buttons in line; this is soo ugly
}))

const Wrapper = styled('div')(({ theme }) => ({
    margin: theme.spacing(1),
    position: 'relative',
}))

const Progress = styled(CircularProgress)(({ theme }) => ({
    color: theme.palette.secondary.main,
    position: 'absolute',
    top: 8,
    left: 8,
    zIndex: 1,
}))

const IntervalButton = styled(Button)(() => ({
    margin: '8px 0',
    borderRadius: '42em',
}))


export interface RefresherProps {
    /** 
     * show throbber if refresh takes longer than the specified threshold;
     * defaults to 500ms. 
     */
    throbberThreshold?: number
    /**
     * an array of refresh intervals; if left undefined, then a default array is
     * applied.
     */
    intervals?: RefresherInterval[]
}

/**
 * A refresher that doesn't stink. This component gives users control over the
 * interval between refreshes, as well as a chance to fire off single-shot
 * on-demand refreshes. Users can switch off automatic refreshing completely. If
 * a refresh takes more than a certain threshold (defaults to 500ms), then a
 * rotating progress indicator appears around the refresh button.
 *
 * This component actually renders two buttons:
 * - on-demand refresh button,
 * - refresh interval selector button, which shows an interval selection menu
 *   when pressed (clicked, touched, ...).
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
const Refresher = ({ throbberThreshold, intervals }: RefresherProps) => {
    const menuId = useId('refreshermenu')

    // Refresh interval and status (is a refresh ongoing?).
    const [refreshInterval, setRefreshInterval] = useAtom(discoveryRefreshIntervalAtom)
    const [refreshing, setRefreshing] = useAtom(discoveryRefreshingAtom)

    // Used for popping up the interval menu.
    const [anchorEl, setAnchorEl] = useState<EventTarget & HTMLElement>()

    // Create the final list of interval values and labels, based on what we
    // were given, or rather, no given.
    intervals = [...(intervals || defaultIntervals)]
        .map(i => ({
            interval: i.interval,
            label: i.label || intervalToLabel(i.interval)
        } as RefresherInterval))

    // User clicks on the auto-refresh button to pop up the associated menu.
    const handleIntervalButtonClick = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget)
    };

    // User selects an auto-refresh interval menu item.
    const handleIntervalMenuChange = (interval: RefresherInterval) => {
        setAnchorEl(undefined)
        setRefreshInterval(interval.interval)
    };

    // User clicks outside the popped up interval menu.
    const handleIntervalMenuClose = () => setAnchorEl(undefined);

    const intervalTitle = refreshInterval !== null
        ? "auto-refresh interval " + intervalToLabel(refreshInterval)
        : "auto-refresh off"

    return (
        <Refreshee>
            <Tooltip title="refresh">
                <Wrapper>
                    <IconButton color="inherit" onClick={() => setRefreshing(true)} size="large"><RefreshIcon /></IconButton>
                    {refreshing &&
                        <Fade
                            in={true}
                            style={{ transitionDelay: `${throbberThreshold || defaultThrobberThreshold}ms` }}
                            unmountOnExit
                        >
                            <Progress size={32} />
                        </Fade>
                    }
                </Wrapper>
            </Tooltip>
            <Tooltip title={intervalTitle}>
                <IntervalButton
                    aria-haspopup="true"
                    aria-controls={menuId}
                    onClick={handleIntervalButtonClick}
                    color="inherit"
                    centerRipple={true}
                >
                    {refreshInterval !== null ? <SyncIcon /> : <SyncDisabledIcon />}
                    <ExpandMoreIcon />
                </IntervalButton>
            </Tooltip>
            <Menu
                id={menuId}
                anchorEl={anchorEl}
                keepMounted
                open={Boolean(anchorEl)}
                onClose={handleIntervalMenuClose}
            >
                {intervals.map(i => {
                    return <MenuItem
                        key={i.interval || -1}
                        value={i.interval || -1}
                        selected={i.interval === refreshInterval}
                        onClick={() => handleIntervalMenuChange(i)}
                    >
                        {i.label}
                    </MenuItem>
                })}
            </Menu>
        </Refreshee>
    );
}

export default Refresher

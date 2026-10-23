// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import '@mui/material/styles'
import type { PaletteColor, SimplePaletteColorOptions } from '@mui/material/styles'

// We augment the existing Material-UI theme with new elements for uniform color
// styling of Ghostwire UI elements beyond the predefined Material UI elements.
// This avoids scattering and potentially duplicating the same color
// configurations all over the various Ghostwire-specific UI elements.
//
// See also:
// https://medium.com/javascript-in-plain-english/extend-material-ui-theme-in-typescript-a462e207131f
declare module '@mui/material/styles' {

    interface Palette {
        address: {
            // stubbornly using IPv6 terminology here :D
            prefix: string,
            iid: string,
        },
        bridgepaper: string,
        capture: string,
        containee: {
            exited: string,
            running: string,
            paused: string,
            pod: string,
            bindmount: string,
            privileged: PaletteColor,
            elevated: PaletteColor,
        },
        operstate: {
            unknown: string,
            up: string,
            down: string,
            lowerlayerdown: string,
            dormant: string,
        },
        routing: {
            selected: PaletteColor,
        },
        wire: {
            down: string,
            hot: string,
            external: string,
            pfvf: string,
            maclvan: string,
            veth: string,
            vxlan: string,
        },
        sched: {
            nice: string // nice nice value color
            notnice: string // not-nice value color
            prio: string // non-0/non-1 prio value color
            relaxed: string // scheduler NORMAL/BATCH/IDLE color
            stressed: string // scheduler FIFO/RR/DEADLINE color
        },
    }

    // allow configuration using `createMuiTheme`
    interface PaletteOptions {
        address?: {
            prefix?: string,
            iid?: string,
        },
        bridgepaper?: string,
        capture?: string,
        containee?: {
            exited?: string,
            running?: string,
            paused?: string,
            pod?: string,
            bindmount?: string,
            privileged?: SimplePaletteColorOptions,
            elevated?: SimplePaletteColorOptions,
        },
        operstate?: {
            unknown?: string,
            up?: string,
            down?: string,
            lowerlayerdown?: string,
            dormant?: string,
        },
        routing?: {
            selected?: SimplePaletteColorOptions,
        },
        wire?: {
            down?: string,
            hot?: string,
            external?: string,
            pfvf?: string,
            maclvan?: string,
            veth?: string,
            vxlan?: string,
        },
        sched?: {
            nice?: string
            notnice?: string
            prio?: string
            relaxedsched?: string
            stressedsched?: string
        },
    }
    
}

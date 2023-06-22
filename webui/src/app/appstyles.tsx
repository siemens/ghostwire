// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { PaletteColor, SimplePaletteColorOptions } from '@mui/material'
import { amber, blue, green, grey, indigo, orange, pink, purple, red, yellow } from '@mui/material/colors'
//import { PaletteColor, SimplePaletteColorOptions } from '@mui/material/styles/createPalette'

import { cloneDeep, merge as mergeDeep } from 'lodash'
import { rgba } from 'utils/rgba'


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
    }
}

// The (basic) light theme parts specific to Ghostwire.
export const gwLightTheme = {
    components: {
        MuiCssBaseline: {
            styleOverrides: {
            },
        },
    },
    palette: {
        background: {
            default: '#fafafa', // restore v4 palette
            paper: '#fff',
        },
        address: {
            prefix: indigo[600],
            iid: purple[700],
        },
        bridgepaper: rgba(blue[900], 0.03),
        capture: blue[900],
        containee: {
            exited: red[800],
            running: green[700],
            paused: amber[600],
            pod: blue[700],
            bindmount: blue[800],
            privileged: {
                main: red.A400,
            },
            elevated: {
                main: yellow[700],
            },
        },
        operstate: {
            unknown: green[700], // sic!
            up: green[700],
            down: red[800],
            lowerlayerdown: red[500],
            dormant: amber[700],
        },
        routing: {
            selected: {
                main: '',
                light: orange[300],
                dark: orange[700],
                contrastText: ''
            },
        },
        wire: {
            down: grey[400],
            hot: pink[900],
            external: `${rgba(indigo[500], 0.5)}`,
            pfvf: orange[300],
            maclvan: '#008b8b',
            veth: '#008800', // Profinepp green ;)
            vxlan: '#b8860b',
        },
    },
}

// The dark theme, based on the light theme.
export const gwDarkTheme = mergeDeep(
    cloneDeep(gwLightTheme),
    {
        palette: {
            background: {
                default: '#303030', // restore v4 palette
                paper: '#424242',
            },
            address: {
                prefix: indigo[200],
                iid: pink[200],
            },
            bridgepaper: rgba(blue[800], 0.03),
            capture: indigo[300],
            containee: {
                pod: blue[500],
                bindmount: blue[700],
            },
            operstate: {
                up: green[500],
                down: red[500],
                lowerlayerdown: red[700],
            },
            wire: {
                down: grey[700],
                external: `${rgba(indigo[300], 0.5)}`,
                pfvf: orange[700],
            },
        },
    }
)

// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { indigo, purple, blue, red, green, amber, yellow, orange, grey, pink, lightGreen } from '@mui/material/colors'
import { rgba } from 'utils/rgba'

// The (basic) light theme parts specific to Ghostwire. Please note that this is
// incomplete, so you normally want to use it with createTheme with the options
// object being at least "{ palette: { mode: 'light' } }".
export const gwLightTheme = {
    components: {
        MuiSelect: {
            defaultProps: {
                variant: 'standard', // MUI v4 default.
            },
        },
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
        sched: {
            nice: lightGreen[700],
            notnice: orange[900],
            prio: red[400],
            relaxedsched: lightGreen[400],
            stressedsched: red[400],
        },
    },
}

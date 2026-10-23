// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { indigo, blue, red, green, orange, grey, pink, lightGreen } from '@mui/material/colors'
import { cloneDeep, merge as mergeDeep } from 'lodash'

import { rgba } from 'utils/rgba'
import { gwLightTheme } from './light'

// The dark theme, based on the light theme. Please note that this is
// incomplete, so you normally want to use it with createTheme with the options
// object being at least "{ palette: { mode: 'dark' } }".
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
            sched: {
                nice: lightGreen[500],
                notnice: orange[500],
            },
        },
    }
)


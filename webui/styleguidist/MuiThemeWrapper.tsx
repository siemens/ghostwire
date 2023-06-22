// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { createTheme, ScopedCssBaseline, ThemeProvider } from '@mui/material'

import { MemoryRouter as Router } from 'react-router'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'


import { gwLightTheme } from 'app/appstyles'
import { DynVarsProvider } from 'components/dynvars'


const lightTheme = createTheme(
    {
        components: {
            MuiScopedCssBaseline: {
                styleOverrides: {
                    root: {
                        fontSize: '0.875rem', // ...go back to typography body2 font size as in MUI v4.
                        lineHeight: 1.43,
                        letterSpacing: '0.01071em',
                    },
                },
            },
        },
        palette: {
            mode: 'light',
            primary: { main: '#009999' },
            secondary: { main: '#ffc400' },
        },
    },
    gwLightTheme,
)


const MuiThemeWrapper = ({ children }) => (
    <ThemeProvider theme={lightTheme}>
        <ScopedCssBaseline>
            <Router>
                <DynVarsProvider value={{enableMonolith: true}}>
                    {children}
                </DynVarsProvider>
            </Router>
        </ScopedCssBaseline>
    </ThemeProvider>
)

export default MuiThemeWrapper

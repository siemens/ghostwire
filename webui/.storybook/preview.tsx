// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Parameters, Preview } from '@storybook/react-vite'

import '@fontsource/roboto/300.css'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

import { MemoryRouter } from "react-router-dom"
import { gwDarkTheme, gwLightTheme } from 'styles/themes'
import { StyledEngineProvider, ThemeProvider, createTheme } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import { IdContext } from 'components/idcontext'

// For themes:Themes exported constant, see:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L20
// and Themes interface type:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L7
import { themes } from 'storybook/theming'

const lightTheme = createTheme(
    { palette: { mode: 'light' } }, gwLightTheme)
const darkTheme = createTheme(
    { palette: { mode: 'dark' } }, gwDarkTheme)

export const parameters: Parameters = {
    options: {
        // sort components by title, but keep the stories in their file order.
        storySort: (a: { title: string }, b: { title: string }) =>
            a.title === b.title ? 0 : a.title.localeCompare(b.title, undefined, { numeric: true }),
    },
    docs: {
        theme: themes.normal, // use same theme as the surrounding parts
        // args often are huge cyclic discovery model graphs: never serialize them.
        source: { type: 'code' },
    },
}

const preview: Preview = {
    decorators: [
        (Story, context) => {
            // support theme changes; see also:
            // https://storybook.js.org/docs/essentials/backgrounds#configuration
            // and
            // https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/backgrounds/types.ts#L21
            //
            // However, the backgrounds.value is undocumented :(
            const isDark = (context.globals.backgrounds?.value || themes.normal.base) === 'dark'
            const theme = isDark ? darkTheme : lightTheme

            // support Story-individual MemoryRouter properties...
            const routerProps = context.parameters.routerProps || 
                { initialEntries: ["/"] }

            return (
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme} >
                        <CssBaseline enableColorScheme />
                        <MemoryRouter {...routerProps}>
                            <div style={{background: theme.palette.background.paper}}>
                                {/* separate DOM element ids of stories sharing the same docs page */}
                                <IdContext>
                                    <Story />
                                </IdContext>
                            </div>
                        </MemoryRouter>
                    </ThemeProvider>
                </StyledEngineProvider>
            )
        },
    ],
}

export default preview
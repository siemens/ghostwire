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
import Box from '@mui/material/Box'
import CssBaseline from '@mui/material/CssBaseline'
import { Markdown } from '@storybook/addon-docs/blocks'
import { IdContext } from 'components/idcontext'

// For themes:Themes exported constant, see:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L20
// and Themes interface type:
// https://github.com/storybookjs/storybook/blob/514890e4cdd4e344c43f3dd03e03e9aaa626b7d9/code/core/src/theming/create.ts#L7
import { ensure, themes, ThemeProvider as SbThemeProvider } from 'storybook/theming'

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

            // docs pages already render story descriptions themselves.
            const description: string | undefined = context.viewMode === 'story'
                ? context.parameters.docs?.description?.story
                : undefined

            return (
                <StyledEngineProvider injectFirst>
                    <ThemeProvider theme={theme} >
                        <CssBaseline enableColorScheme />
                        <MemoryRouter {...routerProps}>
                            {description && (
                                <Box sx={{
                                    mb: 2, pb: 1, borderBottom: 1, borderColor: 'divider',
                                    typography: 'body2', color: 'text.secondary',
                                    '& p': { m: 0 },
                                }}>
                                    <SbThemeProvider theme={ensure(isDark ? themes.dark : themes.light)}>
                                        <Markdown>{description}</Markdown>
                                    </SbThemeProvider>
                                </Box>
                            )}
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
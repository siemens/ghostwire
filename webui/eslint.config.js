import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
    globalIgnores(['dist', 'build', 'styleguide', 'coverage', 'storybook-static', '.yarn']),
    {
        files: ['**/*.{ts,tsx}'],
        extends: [
            js.configs.recommended,
            tseslint.configs.recommended,
            reactHooks.configs.flat.recommended,
            reactRefresh.configs.vite,
        ],
        languageOptions: {
            ecmaVersion: 2020,
            globals: globals.browser,
        },
    },
    {
        // Playwright fixtures call "use", which isn't React's use hook.
        files: ['e2e/**/*.ts', 'playwright*.config.ts'],
        languageOptions: {
            globals: globals.node,
        },
        rules: {
            'react-hooks/rules-of-hooks': 'off',
        },
    },
])
// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import flat from 'core-js/features/array/flat'

import React from 'react'
import { createRoot } from 'react-dom/client'

import './index.css'
import App from './app'

// Import only the necessary Roboto fonts, so they are available "offline"
// without CDN.
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

// HACK: for reasons yet unknown to mankind, the usual direct import of
// 'core-js/features/array/flat' doesn't correctly fix missing Array.flat() on
// some browsers; however, a non-polluting import with explicit pollution then
// works. 
if (Array.flat === undefined) {
	Array.flat = flat
}

// Allow development version to temporarily drop strict mode in order to see
// performance without strict-mode double rendering.
const container = document.getElementById('root');
createRoot(container).render(
	process.env.REACT_APP_UNSTRICT
		? <App />
		: <React.StrictMode><App /></React.StrictMode>
);
// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

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

// Allow development version to temporarily drop strict mode in order to see
// performance without strict-mode double rendering.
const container = document.getElementById('root');
createRoot(container!).render(
	import.meta.env.REACT_APP_UNSTRICT
		? <App />
		: <React.StrictMode><App /></React.StrictMode>
);
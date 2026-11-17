// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { useBrandName } from 'components/branding'

/**
 * Renders the brand name as a simple string: this is either the build-time
 * brand name (defaulting to "Ghostwire") or an optional brand name override
 * via the dynamic variables passed to us by the Ghostwire service.
 */
export const Brand = () => useBrandName()

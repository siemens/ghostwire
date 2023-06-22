// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Tooltip } from '@mui/material'

/**
 * Wrap a component (JSX element) into a tooltip depending on whether the
 * `enableTooltip` parameter is true or not.
 *
 * @param node node wrap with a Material-UI tooltip if `enableTooltip` is
 * `true`.
 * @param tooltip text of tooltip.
 * @param enableTooltip if true, node will be wrapped in tooltip.
 */
export const TooltipWrapper = (node: JSX.Element, tooltip: JSX.Element, enableTooltip: boolean): JSX.Element => {
    return enableTooltip ? <Tooltip title={tooltip}>{node}</Tooltip> : node
}

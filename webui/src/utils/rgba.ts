// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import colorRgba from 'color-rgba'

/**
 * Returns a CSS "rgba(...)" color string given a CSS color string (which
 * optionally might include an alpha value itself) and a separate alpha
 * (transparency) value.
 *
 * @param color color string, such as "#rgb", "#rrggbb", "rgb(r,g,b)", et
 * cetera. Even "rgba(r,g,b,a)" is acceptable.
 * @param alpha alpha value in the range of [0..1].
 */
export const rgba = (color: string, alpha: number) => {
    const [r, g, b, a] = colorRgba(color) || [0, 0, 0, 0]
    return `rgba(${r},${g},${b},${a*alpha})`
}

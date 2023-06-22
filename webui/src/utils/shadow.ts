// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import rgba from 'color-rgba'

// Like Material-Ui's createShadow() (see:
// https://github.com/mui-org/material-ui/blob/master/packages/material-ui/src/styles/shadows.js),
// but working with arbitrary shadow colors.

const shadowKeyUmbraOpacity = 0.2
const shadowKeyPenumbraOpacity = 0.14
const shadowAmbientShadowOpacity = 0.12

/**
 * Returns a string suitable for use as the value of a CSS "box-shadow"
 * property to implement the Material Design shadow design style. Deviating
 * from the one and true spirit of Material Design this function allows
 * creating colored shadows (are these then ... halos?).
 *
 * @param color color in string form, such as '#rgb', '#rrggbb', 'goosegreen',
 * 'rgba(.6,.6,.6,.6)', et cetera.
 * @param px an array of 12 shadow parameters (3 shadows à 4 parameters).
 */
export const createMuiShadow = (color: string, ...px: number[]) => {
    const [r, g, b, ] = rgba(color)
    return [
        `${px[0]}px ${px[1]}px ${px[2]}px ${px[3]}px rgba(${r}, ${g}, ${b}, ${shadowKeyUmbraOpacity})`,
        `${px[4]}px ${px[5]}px ${px[6]}px ${px[7]}px rgba(${r}, ${g}, ${b}, ${shadowKeyPenumbraOpacity})`,
        `${px[8]}px ${px[9]}px ${px[10]}px ${px[11]}px rgba(${r}, ${g}, ${b}, ${shadowAmbientShadowOpacity})`,
    ].join(',')
}

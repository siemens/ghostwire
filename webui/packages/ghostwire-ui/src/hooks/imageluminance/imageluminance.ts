// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useEffect, useState } from 'react'

// Images are downscaled to this size before sampling their pixels.
const SAMPLE_SIZE = 32

// Converts an 8-bit sRGB channel value into linear light.
const linear = (c: number) => {
    const v = c / 255
    return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
}

/**
 * Returns the alpha-weighted mean relative luminance (0..1) of the specified
 * (loaded) image, or undefined if the image cannot be sampled or is fully
 * transparent.
 */
export const imageLuminance = (img: HTMLImageElement): number | undefined => {
    const canvas = document.createElement('canvas')
    canvas.width = SAMPLE_SIZE
    canvas.height = SAMPLE_SIZE
    const ctx = canvas.getContext('2d', { willReadFrequently: true })
    if (!ctx) {
        return undefined
    }
    ctx.drawImage(img, 0, 0, SAMPLE_SIZE, SAMPLE_SIZE)
    let data: Uint8ClampedArray
    try {
        data = ctx.getImageData(0, 0, SAMPLE_SIZE, SAMPLE_SIZE).data
    } catch {
        // tainted canvas due to a cross-origin image without CORS.
        return undefined
    }
    let sum = 0
    let weight = 0
    for (let i = 0; i < data.length; i += 4) {
        const a = data[i + 3] / 255
        if (a === 0) {
            continue
        }
        sum += a * (0.2126 * linear(data[i]) + 0.7152 * linear(data[i + 1]) + 0.0722 * linear(data[i + 2]))
        weight += a
    }
    return weight > 0 ? sum / weight : undefined
}

/**
 * Hook returning the alpha-weighted mean relative luminance (0..1) of the
 * image at the specified URL, ignoring fully transparent pixels. Returns
 * undefined while the image is loading, or if its luminance cannot be
 * determined.
 *
 * @param src image URL, such as a data: URL.
 */
export const useImageLuminance = (src: string | null | undefined) => {
    const [luminance, setLuminance] = useState<{ src: string, value: number | undefined }>()

    useEffect(() => {
        if (!src) {
            return
        }
        let cancelled = false
        const img = new Image()
        img.onload = () => {
            if (!cancelled) {
                setLuminance({ src, value: imageLuminance(img) })
            }
        }
        img.src = src
        return () => { cancelled = true }
    }, [src])

    return luminance && luminance.src === src ? luminance.value : undefined
}

export default useImageLuminance

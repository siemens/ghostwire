// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import scrollIntoView from 'scroll-into-view-if-needed'

export const scrollIdIntoView = (id: string) => {
    const node = document.getElementById(id)
    node && scrollIntoView(node, {
        scrollMode: 'if-needed',
        behavior: 'smooth',
        block: 'start',
        inline: 'start',
    })
}

// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

export const infiniteLifetime = 4294967295

/**
 * Renders the lifetime of an IP address in the format "d hh:mm:ss", or the
 * infinity symbol for infinite address lifetimes. Here, "d" denotes days. And
 * we won't support the Maya calender cycles, ever.
 *
 * @param lifetime address lifetime in seconds.
 */
export const IpAddressLifetime = (lifetime: number) => {
    if (lifetime === infiniteLifetime) {
        return '∞'
    }
    const secs = (lifetime % 60).toString().padStart(2, '0')
    lifetime = (lifetime / 60) >> 0
    const mins = (lifetime % 60).toString().padStart(2, '0')
    lifetime = (lifetime / 60) >> 0
    const hours = (lifetime % 24).toString().padStart(2, '0')
    const days = (lifetime / 24) >> 0
    return `${days}d ${hours}:${mins}:${secs}`
}

export const IpAddressAndPrefix = (addr: string, prefixlen: number) => {
    return (addr.includes('.') && IPv4AddressAndPrefix(addr, prefixlen)) || IPv6AddressAndPrefix(addr, prefixlen)
}

/**
 * Renders an textual IPv4 address into separate prefix and host address
 * parts, based on the given prefix length.
 *
 * @param addr IPv4 address in dotted notation.
 * @param prefixlen length of IPv4 address prefix (0 ≤ prefixlen ≤ 32)
 */
const IPv4AddressAndPrefix = (addr: string, prefixlen: number) => {
    // The textual representation of IPv4 addresses bases on octets, so we can
    // colorize only complete octets, unless someone is keen on using binary
    // format...
    const groups = (prefixlen / 8) >> 0
    const explodedIP = addr.split(".")
    const prefixDigits = explodedIP.slice(0, groups)
    const host = explodedIP.slice(groups)
    const prefix = prefixDigits.length ? prefixDigits.join('.') : ''
    let iid = ''
    if (host.length) {
        if (prefix.length) {
            iid = '.'
        }
        iid += host.join('.')
    }
    return [prefix, iid, prefixlen]
}

/** Trims all leading zeros from a (hex) number string. */
const trimZero = (s: string) => {
    const len = s.length
    for (let idx = 0; idx < len; idx++) {
        if (s.charAt(idx) !== '0') {
            return s.slice(idx)
        }
    }
    // After a complete string of zeros, simply return a single zero ...
    // that's enough!
    return '0'
}

/**
 * Renders an textual IPv46 address into separate prefix and host address
 * parts, based on the given prefix length.
 *
 * @param addr IPv6 address in textual notation, optionally including
 * collapsed regions using a single "::".
 * @param prefixlen length of IPv6 address prefix (0 ≤ prefixlen ≤ 128)
 */
const IPv6AddressAndPrefix = (addr: string, prefixlen: number) => {
    // The textual representation of IPv6 addresses bases on groups of hex
    // nibbles, where each group consists of up to four nibbles. And then,
    // there is the "::" shorthand notation, so we need to "explode" the
    // textual representation first, if necessary.
    let exploded: string // ...without delemiters
    if (addr.includes('::')) {
        const parts = addr.split("::")
        if (parts.length > 2) {
            return ['', addr, prefixlen]
        }
        const before = parts[0].split(':').map(group => group.padStart(4, '0'))
        const after = parts[1].split(':').map(group => group.padStart(4, '0'))
        exploded = before.concat('0000'.repeat(8 - before.length - after.length)).concat(after).join('')
    } else {
        exploded = addr.split(':').map(group => group.padStart(4, '0')).join('')
    }
    let prefixNibbles = (prefixlen / 4) >> 0
    const prefixDigits = exploded.slice(0, prefixNibbles)
    const iidDigits = exploded.slice(prefixNibbles)
    // First part: render the prefix (nibbles), not using zero-group
    // compression, but dropping leading nibble zeros *inside* groups. The
    // prefix nibbles can later be visually differentiated from the IID
    // nibbles, if necessary.
    let collapsedPrefix = false // signals when we collapse groups in the prefix.
    let prefix: string
    if (prefixNibbles) {
        const p = []
        for (let idx = 0; idx < prefixDigits.length; idx += 4) {
            p.push(trimZero(prefixDigits.slice(idx, idx + 4)))
        }
        // Are there leading 0 groups we could collapse?
        let zeroGroupCount = p.findIndex(group => group !== '0')
        if (zeroGroupCount < 0) {
            zeroGroupCount = p.length
        }
        if (zeroGroupCount) {
            // Yes, there are leading 0 groups in the prefix, so we collapse
            // them.
            collapsedPrefix = true
            prefix = '::' + p.slice(zeroGroupCount).join(':')
        } else {
            // No leading 0 groups, but let's see if we could still collapse
            // trailing 0 groups in the prefix...
            const zeroGroupCount = 0 // FIXME: p.slice().reverse().findIndex(group => group !== '0')
            if (zeroGroupCount > 0) {
                collapsedPrefix = true
                prefix = p.slice(0, p.length - zeroGroupCount).join(':') + '::'
            } else {
                // No, no, no; there is nothing to collapse in the prefix, so
                // we're done with the prefix.
                prefix = p.join(':')
            }
        }
    } else {
        prefix = ''
    }
    // As the prefix length may not be a multiple of 16 and thus not fall on a
    // group boundary, we next fill in as many IID nibbles in order to reach
    // the next group boundary, before we start working on IID details and
    // optionally IID group compression.
    let iid = ''
    if (prefixNibbles % 4) {
        iid = iidDigits.slice(0, 4 - (prefixNibbles % 4))
        prefixNibbles += 4 - (prefixNibbles % 4)
    }
    // Now take the remaining IID nibbles, taking zero-group compression
    // ("::") and dropping leading zero nibbles in a group into consideration.
    // We start with grouping the IID nibbles into groups of four nibbles,
    // removing leading 0 nibbles while we're at it.
    const i = []
    for (let idx = prefixNibbles; idx < 32; idx += 4) {
        i.push(trimZero(exploded.slice(idx, idx + 4)))
    }
    if (i.length) {
        // If there wasn't zero group compression in the prefix, then we're
        // free to collapse zero groups in the IID using "::". In this case,
        // we count the number of leading "0000" groups in the IID in order to
        // skip them and only show the remaining groups. Otherwise, we have to
        // show the full IID as normal, but with dropped leading zeros.
        let zeroGroupCount = 0
        if (!collapsedPrefix) {
            zeroGroupCount = i.findIndex(group => group !== '0')
            if (zeroGroupCount < 0) {
                zeroGroupCount = i.length
            }
        }
        if (zeroGroupCount) {
            iid += '::' + i.slice(zeroGroupCount).join(':')
        } else {
            if (prefix.slice(-1) !== ':') {
                iid += ':'
            }
            iid += i.join(':')
        }
    }
    // Phew, finally done...
    return [prefix, iid, prefixlen]
}

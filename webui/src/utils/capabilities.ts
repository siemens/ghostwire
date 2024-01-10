// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import JSBI from 'jsbi'

const caps: { [capno: number]: string } = {
    0: 'CAP_CHOWN',
    1: 'CAP_DAC_OVERRIDE',
    2: 'CAP_DAC_READ_SEARCH',
    3: 'CAP_FOWNER',
    4: 'CAP_FSETID',
    5: 'CAP_KILL',
    6: 'CAP_SETGID',
    7: 'CAP_SETUID',
    8: 'CAP_SETPCAP',
    9: 'CAP_LINUX_IMMUTABLE',
    10: 'CAP_NET_BIND_SERVICE',
    11: 'CAP_NET_BROADCAST',
    12: 'CAP_NET_ADMIN',
    13: 'CAP_NET_RAW',
    14: 'CAP_IPC_LOCK',
    15: 'CAP_IPC_OWNER',
    16: 'CAP_SYS_MODULE',
    17: 'CAP_SYS_RAWIO',
    18: 'CAP_SYS_CHROOT',
    19: 'CAP_SYS_PTRACE',
    20: 'CAP_SYS_PACCT',
    21: 'CAP_SYS_ADMIN',
    22: 'CAP_SYS_BOOT',
    23: 'CAP_SYS_NICE',
    24: 'CAP_SYS_RESOURCE',
    25: 'CAP_SYS_TIME',
    26: 'CAP_SYS_TTY_CONFIG',
    27: 'CAP_MKNOD',
    28: 'CAP_LEASE',
    29: 'CAP_AUDIT_WRITE',
    30: 'CAP_AUDIT_CONTROL',
    31: 'CAP_SETFCAP',
    32: 'CAP_MAC_OVERRIDE',
    33: 'CAP_MAC_ADMIN',
    34: 'CAP_SYSLOG',
    35: 'CAP_WAKE_ALARM',
    36: 'CAP_BLOCK_SUSPEND',
    37: 'CAP_AUDIT_READ',
    38: 'CAP_PERFMON',
    39: 'CAP_BPF',
    40: 'CAP_CHECKPOINT_RESTORE',
}

const bigZero = JSBI.BigInt(0)
const bigOne = JSBI.BigInt(1)

// maps (BigInt) capabilities masks to their corresponding capability names.
const capsmasks: { [capmask: number]: string } = {}

Object.entries(caps).forEach(([capbitno, capname]) => {
    capsmasks[JSBI.leftShift(bigOne, JSBI.BigInt(capbitno)).toString()] = capname
})

// BigInt mask with all known capabilities set.
const knowncapsmask = Object.keys(caps)
    .reduce((mask, capbitno) =>
        JSBI.bitwiseOr(mask, JSBI.leftShift(bigOne, JSBI.BigInt(capbitno)))
        , bigZero)

/**
 * Returns the name of the capability with the specified capbit set.
 *
 * @param capbit a single capability bit (mask); not: bit number.
 */
export const capname = (capbit: JSBI) => {
    return capsmasks[capbit.toString()]
}

// Returns a list of names for those capabilities set for which we don't know
// their name(s).
const unknowncapnames = (capbits: JSBI) => {
    const capnames: string[] = []
    let unknowncapbits = JSBI.bitwiseAnd(capbits, JSBI.bitwiseNot(knowncapsmask))
    for (let capno = 0; JSBI.notEqual(unknowncapbits, bigZero); capno++, unknowncapbits = JSBI.signedRightShift(unknowncapbits, bigOne)) {
        if (JSBI.notEqual(JSBI.bitwiseAnd(unknowncapbits, bigOne), bigZero)) {
            capnames.push(`CAP_${capno}`)
        }
    }
    return capnames
}

/**
 * Returns the names of the capabilities present in the specified capabilities
 * bitset. Capability bits set without a name known to us (yet) get a synthetic
 * CAP_x name, where x is the bit position.
 *
 * @param capbits number representing a bitset of capabilities.
 * @returns 
 */
export const capsnames = (capbits: JSBI) => {
    return Object.entries(capsmasks)
        .reduce((list, [capmask, capname]) =>
            JSBI.notEqual(JSBI.bitwiseAnd(capbits, JSBI.BigInt(capmask)), bigZero) ? [...list, capname] : list,
            [] as string[])
        .concat(unknowncapnames(capbits))
        .sort((a, b) => a.localeCompare(b))
}

/**
 * The set of Docker default capabilities, as documented in
 * https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities.
 * We use an object map here in order to quickly find out if a capability name
 * belongs to the default set, or not.
 */
export const dockerdefaultcaps = {
    'CAP_AUDIT_WRITE': true,
    'CAP_CHOWN': true,
    'CAP_DAC_OVERRIDE': true,
    'CAP_FOWNER': true,
    'CAP_FSETID': true,
    'CAP_KILL': true,
    'CAP_MKNOD': true,
    'CAP_NET_BIND_SERVICE': true,
    'CAP_NET_RAW': true,
    'CAP_SETFCAP': true,
    'CAP_SETGID': true,
    'CAP_SETPCAP': true,
    'CAP_SETUID': true,
    'CAP_SYS_CHROOT': true,
}

// Capabilities bit mask with only the Docker default capabilities bits set.
const dockerdefaultcapsmask = Object.keys(dockerdefaultcaps)
    .reduce((mask, defaultcapname) =>
        JSBI.bitwiseOr(mask,
            JSBI.leftShift(bigOne, JSBI.BigInt(Object.entries(caps).find(([, capname]) => capname === defaultcapname)[0])))
        , bigZero)

export const notDockerDefaultCaps = (caps: JSBI) => {
    return JSBI.notEqual(JSBI.bitwiseAnd(caps, JSBI.bitwiseNot(dockerdefaultcapsmask)), bigZero)
}

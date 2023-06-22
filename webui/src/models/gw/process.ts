// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import JSBI from 'jsbi'

export interface Process {
    /** PID of process */
    pid: number
    /** 
     * "name" of process, derived from process name, command line, et cetera.
     */
    name: string
    /** command line arguments of process. */
    cmdline: string[]
    /** PID namespace identifier (inode number). */
    pidnsid: number
    /** bounded capabilities bitset */
    capbnd: JSBI
}

// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { ForwardedPort } from 'models/gw/forwardedports'
import React, { useEffect, useState } from 'react'

import {
    styled,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    TableSortLabel,
} from '@mui/material'

import { AddressFamily, AddressFamilySet, Containee, containeeDisplayName, isContainer, isPod, netnsId, NetworkNamespace, orderAddresses, PortUser, PrimitiveContainee } from 'models/gw'
import { ContaineeBadge } from 'components/containeebadge'
import { scrollIdIntoView } from 'utils'
import { useNavigate, useMatch } from 'react-router-dom'
import { useContextualId } from 'components/idcontext'
import { Process } from 'components/process'

/**
 * A row in the forwarded port table is showing information for the (unique)
 * combination of one port and one of its port users. We just keep references
 * here to avoid duplicating data all the time, chasing object references
 * instead...
 */
interface PortRow {
    /** forwarded port */
    port: ForwardedPort
    /** 
     * as multiple users (processes) can share the same port (file descriptor),
     * we need to differentiate here and specify the particular user this row
     * corresponds with.
     */
    user: PortUser
    /** arbitrary key for differentiating and identifying rows */
    key: string
}

/**
 * Sorting order of transport port table columns: either ascending or
 * descending.
 */
type Order = 'asc' | 'desc'

/** 
 * For stable sorting we not only take the "value" of a port row (object) into
 * account, but also its original position in case of port row value ties.
 */
type StableRow = [PortRow, number]

/**
 * Stable sort of port rows that keeps the partial order of port rows intact for
 * rows where the sorting order function returns a tie.
 *
 * Please note that the standard ECMA array sort function does not guarantuee
 * stable sorting, only specific implementations might do so. In the end, we
 * have to implement stable sorting ourselves, by simply using the original
 * array indices whenever we need to break the tie when two rows doesn't differ
 * according to the caller-supplied sorting order function.
 *
 * @param arr the array of port table rows.
 * @param orderFn sorting order function comparing two port rows.
 * @param reverse reverse sorting order (defaults to normal sorting order).
 */
const stableSort = (arr: PortRow[], orderFn: (portA: PortRow, portB: PortRow) => number, reverse?: boolean) => {
    const topsyturvy = reverse ? -1 : 1
    return arr.map((row, idx) => [row, idx] as StableRow)
        .sort((rowA, rowB) => ((topsyturvy * orderFn(rowA[0], rowB[0])) || (rowA[1] - rowB[1])))
        .map(row => row[0])
}

/**
 * Describes a specific column (header) of the forwarded port table so we can
 * later sort the rows using the values in this column.
 */
interface ColHeader {
    /** table-unique column ID; this is not a DOM element id. */
    id: string
    /** column header text label. */
    label: string
    /** sorting order function for comparing two rows. */
    orderFn: (portA: PortRow, portB: PortRow) => number
}


interface PortColumnHeaderProps {
    /** column header description object. */
    header: ColHeader
    /**  */
    onRequestSort: (event: React.MouseEvent<unknown>, colHeader: ColHeader) => void
    /** is this column ordered? If true, the column will show a sort arrow */
    ordered: boolean
    /** 
     * if this column is ordered, is it ascending or descending order? This
     * controls the direction in which the sort arrow is pointing, either
     * upwards or downwards.
     */
    order: Order
}

/**
 * Render a transport port table column header with its label text as well as a
 * sort arrow. This arrow will permamently appear (in correct sorting direction)
 * if this column is the current sorting column; otherwise, the arrow will only
 * appear on hovering over the column header cell.
 */
const PortColumnHeader = ({ header, ordered, order, onRequestSort }: PortColumnHeaderProps) => {

    return (
        <TableCell sortDirection={ordered ? order : false}>
            <TableSortLabel
                active={ordered}
                direction={ordered ? order : 'asc'}
                onClick={(ev) => onRequestSort(ev, header)}
            >
                {header.label}
            </TableSortLabel>
        </TableCell>
    )
}

const TransTable = styled(Table)(({ theme }) => ({
    '& .MuiTableCell-root': {
        paddingLeft: theme.spacing(1),
        paddingRight: 0,
    },
}))

const TransHeader = styled(TableHead)(() => ({
    whiteSpace: 'nowrap',
}))

const AddressCell = styled(TableCell)(() => ({
    fontFamily: 'Roboto Mono',
}))

const PortCell = styled(TableCell)(() => ({
    textAlign: 'end',
    fontFamily: 'Roboto Mono',
}))

const UserDetails = styled('span')(({ theme }) => ({
    whiteSpace: 'nowrap',

    '& .MuiSvgIcon-root': {
        marginRight: '0.1em',
        verticalAlign: 'baseline',
        position: 'relative',
        top: '0.2ex',
        color: theme.palette.text.disabled,
    },
}))

/**
 * Returns a stringified user name consisting of the process line, pod,
 * container, et cetera, suitable for sorting.
 */
const userName = (user: PortUser) => {
    const components = []
    const containee = user.containee
    if (containee) {
        if (isContainer(containee) && containee.pod) {
            components.push(containee.pod.name)
        }
        components.push(containeeDisplayName(containee))
    }
    components.push(command(user.cmdline))
    components.push(user.pid.toString())
    return components.join("/")
}

/** Renders a network namespace's clickable containees in case we don't have any
 * serving socket details.
 */
const targetNetns = (netns: NetworkNamespace, onContaineeNavigation?: (containee: PrimitiveContainee) => void) => {
    if (!netns) return ''

    return netns.containers.map(containee =>
        <ContaineeBadge
            key={`${containee.turtleNamespace}-${containee.name}`}
            button
            containee={containee}
            onClick={onContaineeNavigation}
        />)
}

/** Return last path component of first command line element. */
const command = (cmdline: string[]) => {
    const name = cmdline[0].split('/')
    return name[name.length - 1] + ' ' + cmdline.slice(1).join(' ')
}

// The column header descriptions for the forwarded port table; these includes
// the sorting order functions needed in order to correctly sort the table by
// specific column values.
const ForwardedPortTableColumns: ColHeader[] = [
    {
        id: 'proto',
        label: 'Proto',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.protocol.localeCompare(rowB.port.protocol),
    }, {
        id: 'address',
        label: 'Address',
        orderFn: (rowA: PortRow, rowB: PortRow) => orderAddresses(rowA.port.address, rowB.port.address),
    }, {
        id: 'port',
        label: 'Port',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.port - rowB.port.port,
    }, {
        id: 'service',
        label: 'Service',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.servicename.localeCompare(rowB.port.servicename),
    }, {
        id: 'forwardedaddress',
        label: 'Forwarded to',
        orderFn: (rowA: PortRow, rowB: PortRow) => orderAddresses(rowA.port.forwardedAddress, rowB.port.forwardedAddress),
    }, {
        id: 'forwardedport',
        label: 'Port',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.forwardedPort - rowB.port.forwardedPort,
    }, {
        id: 'forwardedservice',
        label: 'Service',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.forwardedServicename.localeCompare(rowB.port.forwardedServicename),
    }, {
        id: 'user',
        label: 'Group · Container · Process',
        orderFn: (rowA: PortRow, rowB: PortRow) => userName(rowA.user).localeCompare(userName(rowB.user))
    },
]

// Sorts the port rows of a transport port table by the specified column(s),
// identified by their ID(s), optionally reversed if an ID is prefixed by "-".
const sortTableRows = (rows: PortRow[], ids: string[]): PortRow[] => {
    // Now sort by the column(s) specified using their column ID(s) ... and
    // finally I have to admit that Javascript is perfectly able to top Python's
    // list incomprehensions in terms of perverse expressiveness!
    return ids.reduce((rows, id) =>
        id ?
            stableSort(
                rows,
                ForwardedPortTableColumns.find(col => col.id === id.replace('-', '')).orderFn,
                id.startsWith('-'))
            : rows,
        rows)
}

/** Helper to generate transport port row keys. */
const portrowkey = (port: ForwardedPort, user: PortUser) => (
    `${port.protocol}-${port.port.toString().padStart(5, '0')}-${port.address.address}`
    + `-${port.forwardedAddress ? port.forwardedPort.toString().padStart(5, '0') : 'x'}-${port.forwardedAddress ? port.forwardedAddress.address : 'x'}`
    + `-${user.pid}`
)

// As the table rows can later be stabely (re)sorted by the user based on
// columns we start with an initial sort in this order of priority:
//
// - original ("host") port,
// - original ("host") address,
// - address forwarded to,
// - transport protocol,
// - port forwarded to.
//
// Please note that due to the way our stable sorting works, we have to
// specify the columns to sort for in reverse order. Specifying a function
// instead of a sorted table value ensures that the sorting is carried out
// only once and not on each component rerender.
const getInitialTableRows = (netns: NetworkNamespace, families: AddressFamilySet) => {
    const keys: { [key: string]: number } = {}
    const fw: { [key: string]: boolean } = {}
    return sortTableRows(
        netns.forwardedPorts
            .filter(port => families.includes(port.address.family))
            .map(port => {
                if (!port.users || port.users.length === 0) {
                    // for some reason we don't know who actually is serving
                    // this forwarded port. So we create a "dummy" user to
                    // signal the situation while ensuring that we still get at
                    // least the forwarded port shown in the table instead of
                    // silently dropping it.
                    const nouser = { pid: 0 } as PortUser
                    return [{ port: port, user: nouser, key: portrowkey(port, nouser) } as PortRow]
                }
                return port.users.map(user => {
                    let key = portrowkey(port, user)
                    let duplicates = keys[key]
                    if (duplicates === undefined) {
                        keys[key] = 0
                    } else {
                        keys[key] = ++duplicates
                        key += '-' + duplicates.toString()
                    }
                    return { port: port, user: user, key: key } as PortRow
                })
            }).flat().map(row => {
                fw[`${row.port.protocol}-${row.port.address.address}-${row.port.port}`] = true
                return row
            }).concat(netns.transportPorts
                .filter(port =>
                    families.includes(port.localAddress.family)
                    && (port.protocol === 'udp' || port.macroState === 'listening'))
                .map(port => port.users.map(user => {
                    const fwport = {
                        protocol: port.protocol,
                        address: port.localAddress,
                        port: port.localPort,
                        servicename: port.localServicename,
                        netns: netns,
                    } as ForwardedPort
                    const key = portrowkey(fwport, user)
                    if (keys[key] === undefined) {
                        if (fw[`${fwport.protocol}-${fwport.address.address}-${fwport.port}`]) {
                            return null
                        }
                        keys[key] = 0
                        return { port: fwport, user: user, key: key } as PortRow
                    }
                    return null
                })).flat().filter(portrow => !!portrow)
            ),
        ['proto', 'address', 'port'])
}

export interface PortsTableProps {
    /** initial and presorted rows of transport port table */
    initialRows: PortRow[]
}

/**
 * Renders the rows of the transport port table and allows user to change the
 * row sorting via the table column headers.
 */
const PortsTable = ({ initialRows }: PortsTableProps) => {
    const navigate = useNavigate()

    const match1 = useMatch('/:base')
    const match2 = useMatch('/:base/:detail')
    const match = (match1 || match2) ? { ...match1, ...match2 } : null

    const domIdBase = useContextualId('')

    const [tableRows, setTableRows] = useState(initialRows)

    // The sorting direction of the currently selected column for sorting.
    const [orderDir, setOrderDir] = useState<Order>('asc')
    // The ID of the column to sort by; '' if the user hasn't yet selected any
    // column for sorting.
    const [orderBy, setOrderBy] = useState<string>('')

    // When the externally supplied row data updates, we need to update our
    // sorted table rows for display too ... of course ... erm, well, yes ;)
    useEffect(() => {
        setTableRows(sortTableRows(initialRows, [orderBy]))
        // Please note that we must NOT depend on "orderBy" here on purpose, as
        // we're otherwise resetting the table rows state each time the user
        // clicks on a table header column to change sorting ... and that's
        // ain't a good idea, sir!
    }, [initialRows])

    // User clicks on a column header in order to sort the table rows by this
    // column. So, let's sort...
    const handleColumnSort = (event: React.MouseEvent<unknown>, colHeader: ColHeader) => {
        if (colHeader.id !== orderBy) {
            // Different row to be sorted: start with an ascending sort of the
            // new column.
            setOrderDir('asc')
            setOrderBy(colHeader.id)
            setTableRows((tablerows) => sortTableRows(tablerows, [colHeader.id]))
        } else {
            // Same column as before, so reverse the sorting order.
            setOrderDir(orderDir === 'asc' ? 'desc' : 'asc')
            setTableRows((tablerows) => sortTableRows(tablerows, [`${orderDir === 'asc' ? '-' : ''}${colHeader.id}`]))
        }
    }

    // When the user clicks on some containee button badge that is listed as a
    // reference in one of the sections, then navigate to it. We have to
    // differentiate between navigating within the full view, which is actually
    // scrollinge into view, and navigating from one detail view to another
    // detail view, where we change the route.
    const handleContaineeNavigation = (containee: Containee) => {
        if (!match) {
            return
        }
        if (match.params['detail']) {
            // change route from existing detail view to new detail view.
            navigate(`/${match.params['base']}/${encodeURIComponent(containee.name)}`)
        }
        // scroll within the overall view.
        scrollIdIntoView(domIdBase + netnsId(
            isPod(containee) ? containee.containers[0].netns
                : containee.netns))
    }


    return (
        <TableContainer
            sx={{ minHeight: 'calc(100% + 2px)', minWidth: 'calc(100% + 2px)' } /* hack: avoid unnecessary scrollbars in screenshots */}
        >
            <TransTable size="small">
                <TransHeader>
                    <TableRow>
                        {ForwardedPortTableColumns.map(column =>
                            <PortColumnHeader
                                key={column.id}
                                header={column}
                                ordered={column.id === orderBy}
                                order={orderDir}
                                onRequestSort={handleColumnSort}
                            />
                        )}
                    </TableRow>
                </TransHeader>
                <TableBody>
                    {tableRows.map(row =>
                        <TableRow key={row.key}>
                            <TableCell>{row.port.protocol.toUpperCase()}</TableCell>
                            <AddressCell>{row.port.address.address}</AddressCell>
                            <PortCell>:{row.port.port}</PortCell>
                            <TableCell>{row.port.servicename}</TableCell>
                            <AddressCell>{(row.port.forwardedAddress && row.port.forwardedAddress.address) || ''}</AddressCell>
                            <PortCell>{(row.port.forwardedAddress && `:${row.port.forwardedPort}`) || ''}</PortCell>
                            <TableCell>{row.port.forwardedServicename}</TableCell>
                            <TableCell>
                                {row.user.pid > 0
                                    ? <UserDetails>
                                        <Process
                                            cmdline={row.user.cmdline}
                                            containee={row.user.containee}
                                            pid={row.user.pid} />
                                    </UserDetails>
                                    : targetNetns(row.port.netns, handleContaineeNavigation)}
                            </TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </TransTable>
        </TableContainer>
    )
}

export interface OpenPortTableProps {
    /** namespace object, including the transport-layer port details. */
    netns: NetworkNamespace
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** hide table completely if empty (after filtering) */
    hideEmpty?: boolean
}

/**
 * Component `OpenPortTable` renders a table with details about the forwarded
 * and "open" (listening TCP respectively open UDP) ports used in a specific
 * network namespace.
 */
export const OpenPortTable = ({ netns, families, hideEmpty }: OpenPortTableProps) => {

    // If not specified by the component user, assume dual stack.
    families = families || [AddressFamily.IPv4, AddressFamily.IPv6]

    // (Re)calculate the initial set of rows based on the network namespace and
    // the address families to show. By calculating the initial rows here in the
    // outer component we avoid endless loops we otherwise end up in when trying
    // to correctly reinitialize the rows whenever the network namespace changes
    // and also change the rows state when the user changes the sorting (order).
    // See also the comments for StackOverflow answer
    // https://stackoverflow.com/a/62982753.
    const portrows = getInitialTableRows(netns, families)

    return ((portrows.length || !hideEmpty) && <PortsTable initialRows={portrows} />) || null
}

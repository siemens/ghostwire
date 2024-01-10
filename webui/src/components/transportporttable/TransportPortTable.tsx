// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

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

import HearingIcon from '@mui/icons-material/Hearing'
import ConnectedIcon from 'icons/portstates/Connected'

import { AddressFamily, AddressFamilySet, containeeDisplayName, isContainer, NetworkNamespace, orderAddresses, orderByState, PortUser, TransportPort } from 'models/gw'
import { Process } from 'components/process'


/**
 * A row in the transport port table is showing information for the (unique)
 * combination of one port and one of its port users. We just keep references
 * here to avoid duplicating data all the time, chasing object references
 * instead...
 */
interface PortRow {
    /** transport port */
    port: TransportPort
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
 * Describes a specific column (header) of the transport port table so we can
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


/** Return last path component of first command line element. */
const command = (cmdline: string[]) => {
    const name = cmdline[0].split('/')
    return name[name.length - 1] + ' ' + cmdline.slice(1).join(' ')
}

// The column header descriptions for the port table; these includes the sorting
// order functions needed in order to correctly sort the table by specific
// column values.
const TransportPortTableColumns: ColHeader[] = [
    {
        id: 'state',
        label: 'St',
        orderFn: (rowA: PortRow, rowB: PortRow) => orderByState(rowA.port, rowB.port),
    }, {
        id: 'proto',
        label: 'Proto',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.protocol.localeCompare(rowB.port.protocol),
    }, {
        id: 'socket',
        label: 'Socket',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.v4mapped === rowB.port.v4mapped ? 0 : rowA.port.v4mapped ? 1 : -1,
    }, {
        id: 'localaddress',
        label: 'Address',
        orderFn: (rowA: PortRow, rowB: PortRow) => orderAddresses(rowA.port.localAddress, rowB.port.localAddress),
    }, {
        id: 'localport',
        label: 'Port',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.localPort - rowB.port.localPort,
    }, {
        id: 'localservice',
        label: 'Service',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.localServicename.localeCompare(rowB.port.localServicename),
    }, {
        id: 'remoteaddress',
        label: 'Remote',
        orderFn: (rowA: PortRow, rowB: PortRow) => orderAddresses(rowA.port.remoteAddress, rowB.port.remoteAddress),
    }, {
        id: 'remoteport',
        label: 'Port',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.remotePort - rowB.port.remotePort,
    }, {
        id: 'remoteservice',
        label: 'Service',
        orderFn: (rowA: PortRow, rowB: PortRow) => rowA.port.remoteServicename.localeCompare(rowB.port.remoteServicename),
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
                TransportPortTableColumns.find(col => col.id === id.replace('-', '')).orderFn,
                id.startsWith('-'))
            : rows,
        rows)
}

/** Helper to generate transport port row keys. */
const portrowkey = (port: TransportPort, user: PortUser) => (
    `${port.protocol}${port.macroState === 'listening' ? '1' : '2'}`
    + `-${port.localPort.toString().padStart(5, '0')}-${port.localAddress.address}`
    + `-${port.remotePort.toString().padStart(5, '0')}-${port.remoteAddress.address}`
    + `-${user.pid}`
)

// As the table rows can later be stabely (re)sorted by the user based on
// columns we start with an initial sort in this order of priority:
//
// - local port,
// - transport protocol,
// - (reverse) connection state -- to show listening ports before connected ports,
// - local address,
// - remote address,
// - remote port.
//
// Please note that due to the way our stable sorting works, we have to
// specify the columns to sort for in reverse order. Specifying a function
// instead of a sorted table value ensures that the sorting is carried out
// only once and not on each component rerender.
const getInitialTableRows = (netns: NetworkNamespace, families: AddressFamilySet) => {
    const keys: { [key: string]: number } = {}
    return sortTableRows(
        netns.transportPorts
            .filter(port => families.includes(port.localAddress.family))
            .map(port => port.users.map(user => {
                let key = portrowkey(port, user)
                let duplicates = keys[key]
                if (duplicates === undefined) {
                    keys[key] = 0
                } else {
                    keys[key] = ++duplicates
                    key += '-' + duplicates.toString()
                }
                return { port: port, user: user, key: key } as PortRow
            })).flat(),
        ['remoteport', 'remoteaddress', 'localaddress', '-state', 'proto', 'localport'])
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

    return (
        <TableContainer
            sx={{ minHeight: 'calc(100% + 2px)', minWidth: 'calc(100% + 2px)' } /* hack: avoid unnecessary scrollbars in screenshots */}
        >
            <TransTable size="small">
                <TransHeader>
                    <TableRow>
                        {TransportPortTableColumns.map(column =>
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
                            <TableCell>{row.port.macroState === 'connected' ? <ConnectedIcon fontSize="inherit" /> : <HearingIcon fontSize="inherit" />}</TableCell>
                            <TableCell>{row.port.protocol.toUpperCase()}</TableCell>
                            <TableCell>{row.port.v4mapped ? 'IPv6' : ''}</TableCell>
                            <AddressCell>{row.port.localAddress.address}</AddressCell>
                            <PortCell>:{row.port.localPort}</PortCell>
                            <TableCell>{row.port.localServicename}</TableCell>
                            <AddressCell>{(row.port.remotePort && row.port.remoteAddress.address) || ''}</AddressCell>
                            <PortCell>{(row.port.remotePort && `:${row.port.remotePort}`) || ''}</PortCell>
                            <TableCell>{row.port.remoteServicename}</TableCell>
                            <TableCell>
                                <Process cmdline={row.user.cmdline} containee={row.user.containee} pid={row.user.pid} />
                            </TableCell>
                        </TableRow>
                    )}
                </TableBody>
            </TransTable>
        </TableContainer>
    )
}

export interface TransportPortTableProps {
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
 * Component `TransportPortTable` renders a table with details about the
 * transport-layer ports used in a specific network namespace.
 *
 * This component's name on purpose stutters to such an extreme extend that it
 * will end the Go-phers' world domination by making them cry endlessly.
 */
export const TransportPortTable = ({ netns, families, hideEmpty }: TransportPortTableProps) => {

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

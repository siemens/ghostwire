// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { ReactNode } from 'react'

import ClearIcon from '@mui/icons-material/Clear'

import { AddressFamily, AddressFamilySet, PrimitiveContainee, containeeDescription, Container, ContainerState, containerStateString, isSandbox, containeeDisplayName, isContainer, orderAddresses, hiddenLabel } from 'models/gw'
import { Address } from 'components/address'
import StoppedState from 'icons/containeestates/StoppedState'
import PausedState from 'icons/containeestates/PausedState'
import RunningState from 'icons/containeestates/RunningState'
import { ContaineeIcon } from 'utils/containeeicon'
import { useAtom } from 'jotai'
import { showNamespaceIdsAtom } from 'views/settings'
import { capsnames } from 'utils/capabilities'
import Capability from 'components/capability/Capability'
import { styled } from '@mui/material';


// Set the things applying to all child elements within a single containee
// detail component.
const Containees = styled('div')(({ theme }) => ({
    marginBottom: theme.spacing(1),
    fontWeight: 'bold',

    '& > .MuiSvgIcon-root': { verticalAlign: 'middle' },
    '& > .MuiSvgIcon-root.exited': { color: theme.palette.containee.exited },
    '& > .MuiSvgIcon-root.running': { color: theme.palette.containee.running },
    '& > .MuiSvgIcon-root.restarted': { color: theme.palette.containee.running },
    '& > .MuiSvgIcon-root.paused': { color: theme.palette.containee.paused },
}))

// container/process details are a grid (table) of key-value pairs; even
// if some values are arrays consisting of multiple values themselves, but
// we won't notice at this level.
const Details = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'minmax(auto, max-content) minmax(50%, auto)', // https://stackoverflow.com/a/62163763
    columnGap: theme.spacing(2),
    rowGap: theme.spacing(1),

    marginTop: theme.spacing(1),
    marginLeft: theme.spacing(4),
}))

const Property = styled('div')(({ theme }) => ({
    gridColumn: '1 / 2',
    color: theme.palette.text.secondary,
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    overflowWrap: 'break-word',
}))

const Value = styled('div')(() => ({
    gridColumn: '2 / 3',
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    overflowWrap: 'break-word',

    '& > .MuiSvgIcon-root': { verticalAlign: 'middle' },
}))

// The command line of a containee process is separated into individual
// elements and we use the ::before and ::after pseudo elements to place
// bottom corner markers at the beginning and end of each individual
// element. This makes embedded spaces easily spottable so that they can be
// differentiated from the boundaries of individual command line elements.
const CmdlineItem = styled('span')(({ theme }) => ({
    position: 'relative',
    paddingLeft: '2px',
    marginRight: '0.15em',
    '&::before': {
        position: 'absolute',
        content: '"⌜"',
        color: theme.palette.divider,
        top: '-0.6ex',
        left: '-0.15em',
        width: theme.spacing(1),
        height: theme.spacing(1),
    },
    '&::after': {
        position: 'relative',
        content: '"⌟"',
        color: theme.palette.divider,
        bottom: '-0.4ex',
        width: theme.spacing(1),
        height: theme.spacing(1),
        right: '0.2em',
    },
}))

// Host name-address pairs are forming a sub grid (table) inside a value
// grid item; the same goes for container label key-value pairs...
const KeyValList = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'minmax(auto, max-content) minmax(50%, auto)',
    columnGap: theme.spacing(2),
    minHeight: '24px', // ensures consistent height when no icon in value.
}))

const KvKey = styled('div')(() => ({
    gridColumn: '1 / 2',
    alignSelf: 'baseline',
    minHeight: '24px', // ensures consistent height when no icon in value.
    overflowWrap: 'anywhere',
}))

const KvValue = styled('div')(() => ({
    gridColumn: '2 / 3',
    alignSelf: 'baseline',
    minHeight: '24px', // ensures consistent height when no icon in value.
    overflowWrap: 'break-word',
}))

const DNSServers = styled('div')(() => ({
    display: 'grid',
    gridTemplateColumns: 'auto',
    minHeight: '24px', // ensures consistent height.
}))

const DNSServerAddress = styled('div')(() => ({
    gridColumn: '1 / 2',
    minHeight: '24px', // ensures consistent height.
}))

const Capabilities = styled('div')(({ theme }) => ({
    display: 'grid',
    // As we cannot use max-content with repeat() we need to specify a fixed
    // maximum column width: this should fit with the choosen Roboto
    // thickness and maximum capability name.
    gridTemplateColumns: 'repeat(auto-fill, 14em)',
    columnGap: theme.spacing(2),
}))



const containeeStatusIcons = {
    [ContainerState.Exited]: StoppedState,
    [ContainerState.Paused]: PausedState,
    [ContainerState.Restarted]: RunningState,
    [ContainerState.Running]: RunningState,
}

export interface ContaineeDetailsProps {
    /** the containee object for which to render detail information. */
    containee: PrimitiveContainee
    /** CSS class name(s) to apply to the ContaineeDetail root element. */
    className?: string
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
}

/**
 * `ContaineeDetails` shows the detail information about a particular
 * "containee", that is, about a container or process (but not bind-mounts,
 * where there are no useful details such as DNS configuration).
 *
 * Details are:
 * - the (ealdorman) process, with PID and command line. Please note how the
 *   individual command line arguments get visually identified in an
 *   unobtrusive manner using small gray corner adornments at the bottom.
 * - the DNS/name resolution related configuration of this process, especially
 *   UTS hostname, /etc/hostname, the host aliases from /etc/hosts, ...
 *
 * Please note that the host name to address mappings are sorted by the host
 * names.
 */
export const ContaineeDetails = ({ containee, families: fams, className }: ContaineeDetailsProps) => {
    const [showNamespaceIds] = useAtom(showNamespaceIdsAtom)

    const families = fams || [AddressFamily.IPv4, AddressFamily.IPv6]

    // Render a single property with value row in the property grid.
    let row = 0
    const prop = (name: string, value: ReactNode) => {
        if (!value) {
            value = <ClearIcon fontSize="inherit" color="disabled" />
        }
        row++
        return [
            <Property key={`k${row}`} style={{ gridRow: `${row}/${row + 1}` }}>
                {name}
            </Property>,
            <Value key={`v${row}`} style={{ gridRow: `${row}/${row + 1}` }}>
                {value}
            </Value>
        ]
    }

    const cntr = containee as Container
    if (!cntr.dns) {
        return null
    }

    // Render a (sorted) grid of container labels, if available. Do not render
    // any pseudo labels that the Ghostwire backend attached and that are not
    // user labels.
    const containerLabelKeys = ((cntr.labels && Object.keys(cntr.labels)) || [])
        .filter(key => !hiddenLabel(key))
    const containerLabels = (containerLabelKeys.length && <KeyValList>
        {containerLabelKeys.sort((keyA, keyB) => keyA.localeCompare(keyB))
            .map(key => [
                <KvKey key={key}>{key}:</KvKey>,
                <KvValue key={`${key}-value`}>{cntr.labels[key]}</KvValue>
            ])}
    </KeyValList>) || null

    // Render a (sorted) grid of host-address mappings (originating from
    // /etc/hosts). This list is subject to the address family filtering. As
    // the same host name cannot legitimately appear twice in the /etc/hosts
    // files, we simply sort by host name and not caring about the associated
    // address in any way.
    const etcHosts = <KeyValList>
        {cntr.dns.etcHosts
            .filter(namedHost => families.includes(namedHost.address.family))
            .sort((namedhostA, namedhostB) => namedhostA.name.localeCompare(namedhostB.name))
            .map(namedhost => [
                <KvKey key={namedhost.name}>{namedhost.name}</KvKey>,
                <KvValue key={`${namedhost.name}-${namedhost.address.address}`}>
                    <Address address={namedhost.address} plain familyicon />
                </KvValue>
            ])}
    </KeyValList>

    // Render a list of DNS server IP addresses as a single column grid.
    const dnsServers = <DNSServers>
        {cntr.dns.nameservers
            .filter(addr => families.includes(addr.family))
            .sort((addrA, addrB) => orderAddresses(addrA, addrB))
            .map(addr =>
                <DNSServerAddress key={addr.address}>
                    <Address address={addr} plain familyicon />
                </DNSServerAddress>
            )}
    </DNSServers>

    const Icon = containeeStatusIcons[cntr.state] || RunningState
    const CIcon = ContaineeIcon(containee, true)
    const displayName = containeeDisplayName(containee)
    const plaincaps = !isContainer(containee)
    const caps = (cntr.ealdorman &&
        <Capabilities>
            {capsnames(cntr.ealdorman.capbnd).map(capname =>
                <Capability key={capname} name={capname} plain={plaincaps} />)}
        </Capabilities>)

    return (
        <div className={className}>
            <Containees>
                <Icon className={containerStateString(cntr.state) || 'running'} />
                {displayName}
            </Containees>
            <Details>
                {isContainer(cntr) && cntr.id !== displayName && prop('container ID', cntr.id)}
                {prop('type', <>
                    <CIcon fontSize="inherit" color="disabled" />
                    &nbsp;{containeeDescription(containee)}
                </>)}
                {!isSandbox(containee) && [
                    prop('command', cntr.ealdorman.cmdline.map((cmditem, idx) =>
                        <CmdlineItem key={idx}>{cmditem}</CmdlineItem>)),
                    prop('PID', cntr.ealdorman.pid),
                    showNamespaceIds && prop('PID namespace ID', cntr.ealdorman.pidnsid),
                    prop('bounding caps', caps),
                ]}
                {cntr.labels && prop('container labels', containerLabels)}
                {prop('UTS hostname', cntr.dns.utsHostname)}
                {prop('etc hostname', cntr.dns.etcHostname)}
                {prop('etc domainname', cntr.dns.etcDomainname)}
                {prop('DNS servers', dnsServers)}
                {prop('DNS search list', cntr.dns.searchlist && cntr.dns.searchlist.join(', '))}
                {prop('etc hosts', etcHosts)}
            </Details>
        </div>
    )
}

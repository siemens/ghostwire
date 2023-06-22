// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useEffect, useState } from 'react'

import { useNavigate, useMatch } from 'react-router-dom'
import clsx from 'clsx'

import useWebSocket, { ReadyState } from 'react-use-websocket'

import { Button, styled, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Typography } from '@mui/material'
import { Label as LabelIcon, HourglassTop, LocalHospital, QuestionMark, Verified, VerifiedOutlined } from '@mui/icons-material'
import { Service, shareContainers, sortServices } from 'utils/neighborhood'
import { ContaineeBadge } from 'components/containeebadge'
import { rgba } from 'utils/rgba'
import { AddressFamily, Containee, Container, IpAddress, isContainer, orderAddresses } from 'models/gw'
import { basename } from 'utils/basename'
import { Address } from 'components/address'


const ServiceTable = styled(TableContainer)(({ theme }) => ({
    '& .MuiTableCell-root': {
        paddingLeft: theme.spacing(1),
        paddingRight: theme.spacing(1),
    },

    '& .MuiTableCell-root.middlecolumn': {
        paddingLeft: theme.spacing(2),
        paddingRight: theme.spacing(2),
    },

    '& .dnsname': {
        paddingLeft: `calc(14px + 0.1em)`,
        textIndent: `calc(-14px - 0.1em)`,
        whiteSpace: 'nowrap',
    },

    '& .dnsname:not(.double-label)': {
        color: theme.palette.text.secondary,
    },

    '& .qa': {
        paddingLeft: `calc(14px + 0.1em)`,
    },

    '& .qa + .dnsname': {
        paddingTop: theme.spacing(1),
    },

    // display information about our own neighborhood service in a less
    // prominent way, compared to all other neighborhood service information.
    '& .MuiTableRow-root.itsme .MuiTableCell-root': {
        background: rgba(theme.palette.text.disabled, 0.05),
    },
    '& .itsme .MuiTableCell-root': {
        color: theme.palette.text.secondary,
    },

    '& .MuiSvgIcon-root.dns': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.15ex',
        color: rgba(theme.palette.text.primary, 0.1),
        marginRight: '0.1em',
    },
}))

interface QualifiedServiceAddress {
    address: IpAddress
    quality: string
}

interface FullyQualifiedServiceAddresses {
    [fqdn: string]: { [address: string]: QualifiedServiceAddress }
}


const QA = ({ className, qa }: { className?: string, qa: QualifiedServiceAddress }) => {
    var qual
    switch (qa.quality) {
        case 'unverified':
            qual = <QuestionMark className="quality" fontSize="small" />
            break
        case 'verifying':
            qual = <HourglassTop className="quality" fontSize="small" />
            break
        case 'verified':
            qual = <Verified className="quality" fontSize="small" />
            break
        case 'invalid':
            qual = <LocalHospital className="quality" fontSize="small" />
            break
    }

    return <div key={qa.address.address} className={clsx(className, qa.quality)}>
        {qual}
        <Address
            address={qa.address}
            familyicon={true}
            plain={true}
        />
    </div>
}

const QualAddress = styled(QA)(({ theme }) => ({
    '& .MuiSvgIcon-root.quality': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '-0.15ex',
    },
    '&.unverified .MuiSvgIcon-root.quality': {
        color: theme.palette.text.secondary,
    },
    '&.verifying .MuiSvgIcon-root.quality': {
        color: theme.palette.info.main,
        animation: 'spin 1s linear infinite',
    },
    '&.verified .MuiSvgIcon-root.quality': {
        color: theme.palette.success.main,
    },
    '&.invalid .MuiSvgIcon-root.quality': {
        color: theme.palette.error.main,
    },
    '@keyframes spin': {
        '0%': {
            transform: 'rotate(-360deg)',
        },
        '100%': {
            transform: 'rotate(0deg)',
        },
    },
}))

export interface NeighboorhoodProps {
    services: Service[]
    seenby?: Containee[]
}

/**
 * Render the services in the neighborhood of the containers with a specific
 * network namespace: that is, the services available on Docker those networks
 * to which a particular container(s) with a particular network namespace is
 * connected to. Services do not only include those services fulfilled by
 * potentially multiple horizontally scaled containers, but also the
 * individually addressable containers as a non-scaling service.
 */
export const Neighborhood = ({ services, seenby }: NeighboorhoodProps) => {
    seenby = (seenby || []).filter(cntr => isContainer(cntr))

    const navigate = useNavigate()

    const match1 = useMatch('/:base')
    const match2 = useMatch('/:base/:detail')
    const match = (match1 || match2) ? { ...match1, ...match2 } : null

    var url = (window.location.protocol === "https:" ? "wss://" : "ws://")
        + window.location.host + basename + "/mobydig"

    const [fqdnAddrs, setFqdnAddrs] = useState({} as FullyQualifiedServiceAddresses)

    const seenbyName = (seenby.find(cntr => cntr !== undefined) as Container).name
    const [runCheck, setRunCheck] = useState(false)
    const { lastJsonMessage, readyState } = useWebSocket(url, {
        queryParams: { target: seenbyName },
        shouldReconnect: () => false,
    }, runCheck)

    useEffect(() => {
        console.log("websocket ready state", readyState)
        if (readyState === ReadyState.CLOSED || readyState === ReadyState.UNINSTANTIATED) {
            setRunCheck(false)
        }
    }, [readyState])

    useEffect(() => {
        if (!lastJsonMessage) {
            return
        }
        const fqdn = lastJsonMessage['fqdn'].slice(0, -1)
        const addr = lastJsonMessage['address']
        if (!!addr) {
            setFqdnAddrs((fqdnAddrs) => {
                const updatedFqdnAddrs = {
                    ...fqdnAddrs,
                    [fqdn]: { ...fqdnAddrs[fqdn] },
                }
                const fam = addr.includes(':') ? AddressFamily.IPv6 : AddressFamily.IPv4
                updatedFqdnAddrs[fqdn][addr] = {
                    address: {
                        address: addr,
                        family: fam,
                        prefixlen: fam === AddressFamily.IPv6 ? 128 : 32,
                    },
                    quality: lastJsonMessage['quality'],
                } as QualifiedServiceAddress
                return updatedFqdnAddrs
            })
        }
    }, [lastJsonMessage])

    const handleCheck = () => {
        setFqdnAddrs({})
        setRunCheck(true)
    }

    const handleContaineeNavigation = (containee: Containee) => {
        if (!match || !match.params['detail']) {
            return
        }
        // change route from existing detail view to new detail view.
        navigate(`/${match.params['base']}/${encodeURIComponent(containee.name)}`)
    }

    return (services && services.length) ?
        <div>
            <Button
                variant="outlined"
                startIcon={<VerifiedOutlined />}
                disabled={runCheck}
                onClick={handleCheck}
            >Check</Button>
            <ServiceTable
                sx={{ minHeight: 'calc(100% + 2px)', minWidth: 'calc(100% + 2px)' } /* hack: avoid unnecessary scrollbars in screenshots */}
            >
                <Table size="small">
                    <TableHead>
                        <TableRow>
                            <TableCell key="container">Neighbor Container</TableCell>
                            <TableCell key="servicenames" className="middlecolumn">Service DNS Names</TableCell>
                            <TableCell key="containernames">Container DNS Names</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {services.sort((a, b) => sortServices(a, b))
                            .map((service, idx) => {
                                const itsme = shareContainers(service.containers, seenby as Container[])
                                const tlds = ['', ...service.networks].sort((a, b) => a.localeCompare(b))
                                return service.containers
                                    .sort((a, b) => a.name.localeCompare(b.name))
                                    .map((cntr, idx) =>
                                        <TableRow key={`${service.name}-${cntr.name}`} className={clsx(itsme && 'itsme')}>
                                            <TableCell key={cntr.name}>
                                                <ContaineeBadge
                                                    containee={cntr}
                                                    button={!itsme}
                                                    onClick={handleContaineeNavigation}
                                                />
                                            </TableCell>

                                            {idx === 0 && <TableCell key=":servicenames"
                                                rowSpan={service.containers.length}
                                                className="middlecolumn"
                                            >
                                                {!!service.name && tlds.map(tld => {
                                                    const fqdn = service.name + (tld ? '.' + tld : '')
                                                    return <div key={fqdn}>
                                                        <div key={fqdn} className={clsx('dnsname', !!tld && 'double-label')}>
                                                            <LabelIcon className="dns" fontSize="inherit" />{fqdn}
                                                        </div>
                                                        {!!fqdnAddrs[fqdn] && Object.values(fqdnAddrs[fqdn])
                                                            .sort((qa, qb) => orderAddresses(qa.address, qb.address))
                                                            .map(qa =>
                                                                <QualAddress className="qa" key={qa.address.address} qa={qa} />)}
                                                    </div>
                                                })}
                                            </TableCell>}

                                            <TableCell key={`dns-names:${cntr.name}`}>
                                                {tlds.map(tld => {
                                                    const fqdn = cntr.name + (tld ? '.' + tld : '')
                                                    return <div key={fqdn}>
                                                        <div key={fqdn} className={clsx('dnsname', !!tld && 'double-label')}>
                                                            <LabelIcon className="dns" fontSize="inherit" />{fqdn}
                                                        </div>
                                                        {!!fqdnAddrs[fqdn] && Object.values(fqdnAddrs[fqdn])
                                                            .sort((qa, qb) => orderAddresses(qa.address, qb.address))
                                                            .map(qa =>
                                                                <QualAddress className="qa" key={qa.address.address} qa={qa} />)}
                                                    </div>
                                                })}
                                            </TableCell>
                                        </TableRow>)
                            })}
                    </TableBody>
                </Table>
            </ServiceTable>
        </div>
        : <Typography variant="body2" color="textSecondary" paragraph />
}

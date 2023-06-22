// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { styled, Tooltip } from '@mui/material'

import { ContaineeDetails } from 'components/containeedetails'
import { AddressFamilySet, containeesOfNetns, isPod, isSandbox, NetworkNamespace, sortContaineesByName } from 'models/gw'
import { ContaineeIcon } from 'utils/containeeicon'


const Containees = styled('div')(({ theme }) => ({
    '& > * + *': {
        marginTop: theme.spacing(2),
    },
}))

const PodContainees = styled(Containees)(({ theme }) => ({
    borderWidth: 1,
    borderStyle: 'dashed',
    borderColor: theme.palette.divider,
    borderRadius: theme.spacing(1),

    '& > :first-child': {
        marginLeft: 0,
    },
    '& > *': {
        marginLeft: theme.spacing(4),
    },
    '& > * + *': {
        marginTop: theme.spacing(1),
    },

    '& .sandbox': {
        color: theme.palette.text.disabled,
    }
}))

const PodIcon = styled('span')(({ theme }) => ({
    verticalAlign: 'middle',
    color: theme.palette.containee.pod,
}))

export interface DetailedContaineesProps {
    /** 
     * network namespace object (with tons of information thanks to the
     * Ghostwire discovery engine *snicker*). 
     */
    netns: NetworkNamespace
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** show sandbox containers? */
    showSandbox?: boolean
}

/**
 * Component `DetailedContainees` shows the details of all containees of a
 * particular network namespace. Passive containees, that is, bind-mount
 * containees, are *not* shown, only containers and stand-alone processes.
 *
 * The rationale for not showing bind-mount containees is that they don't have
 * any processes and thus the bind mount is necessary to keep a network
 * namespace alive at all. But without any processes there are also is no DNS
 * client configuration, no UTS namespace, et cetera.
 */
export const DetailedContainees = ({ netns, families, showSandbox }: DetailedContaineesProps) => {
    return (<Containees>
        {containeesOfNetns(netns)
            .filter(containee => !isSandbox(containee))
            .sort(sortContaineesByName)
            .map(containee => isPod(containee) ?
                <PodContainees key={containee.name}>
                    <PodIcon as={ContaineeIcon(containee)} />{containee.name}
                    {containee.containers
                        .filter(container => !container.sandbox || showSandbox)
                        .sort(sortContaineesByName)
                        .map(containee =>
                            <ContaineeDetails
                                key={containee.name}
                                containee={containee}
                                families={families}
                            />
                        )}
                    {containee.containers.find(container => container.sandbox && !showSandbox) &&
                        <div className="sandbox">
                            <Tooltip title="sandbox container suppressed in user settings">
                                <span>[sandbox]</span>
                            </Tooltip>
                        </div>
                    }
                </PodContainees>
                : <ContaineeDetails
                    key={containee.name}
                    containee={containee}
                    families={families}
                />
            )
        }
    </Containees>)
}

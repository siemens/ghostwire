// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { styled } from '@mui/material'

import { PrimitiveContainee, NetworkNamespace, sortContaineesByName, containeesOfNetns, isPod, Containee } from 'models/gw'
import { ContaineeBadge } from 'components/containeebadge'


// default maximum number of containee badges to show when the limit property
// of the NamespaceContainees component is left unspecified.
const defaultBadgeLimit = 3

const Containees = styled('span')(({ theme }) => ({
    verticalAlign: 'top',
    marginLeft: theme.spacing(1),

    '& .separator': {
        color: theme.palette.text.disabled,
        fontSize: '90%',
        paddingLeft: '0.2em',
        paddingRight: '0.2em',
    }
}))


export interface NamespaceContaineesProps {
    /** 
     * network namespace object with information about the containees attached
     * to the namespace, such as containers, stand-alone processes and
     * bind-mounts. 
     */
    netns: NetworkNamespace
    /** maximum number of containees shown. */
    limit?: number
    /**
     * callback triggering when the user wants to navigate to one of the
     * containees of a network interface.
     */
    onNavigation?: (containee: PrimitiveContainee) => void
}

/**
 * Component `NamespaceContainees` renders the set of boxed entity badges
 * (`ContaineeBadge`s) for the given network namespace. In case there are
 * multiple containees inside the same namespace, this component limits the
 * number of badges shown. This avoiding "container snakes" on systems with
 * busy (host) network namespaces, such as certain Kubernetes configurations
 * with lots of system containers.
 *
 * Please note that the containee (badges) are always rendered as buttons.
 */
export const NamespaceContainees = ({ netns, limit, onNavigation }: NamespaceContaineesProps) => {
    const containees = containeesOfNetns(netns)
        .sort(sortContaineesByName)
        .slice(0, limit || defaultBadgeLimit)

    // Did we need to cut the list of boxed entities short? Then later show the
    // usual ellipsis indication when rendering the badge list...
    const optionalEllipsis = (netns.containers.length > containees.length &&
        <span className="separator">…</span>) || ''

    // trigger the callback for navigating to a specific containee.
    const handleBadgeClick = (containee: PrimitiveContainee) => {
        if (onNavigation) {
            onNavigation(containee)
        }
    }

    return (
        <Containees>
            {containees.map((containee, idx) => (
                <span
                    key={`${isPod(containee) ? containee.containers[0].netns : containee.netns}-${containee.name}`}
                >
                    {idx > 0 && <span className="separator">+</span>}
                    <ContaineeBadge
                        key={containee.name}
                        containee={containee}
                        button
                        onClick={handleBadgeClick as (_: Containee) => void}
                    />
                </span>))}{optionalEllipsis}
        </Containees>
    )
}

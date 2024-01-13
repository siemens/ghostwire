// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { Link as RouterLink, LinkProps as RouterLinkProps, useMatch } from 'react-router-dom'
import { Avatar, List, ListItem, ListItemAvatar, ListItemButton, ListSubheader, styled, Typography } from '@mui/material'

import { Containee, containeeState, Container, ContainerState, containerStateString, emptyNetns, isContainer, isElevatedContainer, isPod, isPodContainer, isPrivilegedContainer, netnsId, NetworkNamespaces, Pod, PodFlavors, sortContaineesByName } from 'models/gw'
import { ContaineeIcon } from 'utils/containeeicon'
import { useContextualId } from 'components/idcontext'
import PrivilegedIcon from 'icons/containeestates/Privileged'
import CapableIcon from 'icons/containeestates/Capable'


const ContaineeItem = styled(ListItemButton)<RouterLinkProps>(({ theme }) => ({
    '&.privileged > .MuiSvgIcon-root': {
        fill: theme.palette.containee.privileged.main,
    },
    '&.elevated > .MuiSvgIcon-root': {
        fill: theme.palette.containee.elevated.main,
    },
})) as unknown as typeof ListItemButton

const StatefulAvatar = styled(Avatar)(({ theme }) => ({
    width: theme.spacing(4),
    height: theme.spacing(4),

    backgroundColor: theme.palette.containee.bindmount,

    '&.MuiAvatar-colorDefault': {
        color: theme.palette.mode === 'light' ? theme.palette.background.default : theme.palette.text.primary,
    },

    [`&.${containerStateString(ContainerState.Running)}`]: {
        color: theme.palette.containee.running,
        backgroundColor: "inherit",
        borderColor: theme.palette.containee.running,
        borderWidth: 2,
        borderStyle: "solid",
    },
    [`&.${containerStateString(ContainerState.Restarted)}`]: {
        color: theme.palette.containee.running,
        backgroundColor: "inherit",
        borderColor: theme.palette.containee.running,
        borderWidth: 2,
        borderStyle: "solid",
    },
    [`&.${containerStateString(ContainerState.Paused)}`]: {
        color: theme.palette.containee.paused,
        backgroundColor: "inherit",
        borderColor: theme.palette.containee.paused,
        borderWidth: 2,
        borderStyle: "solid",
    },
    [`&.${containerStateString(ContainerState.Exited)}`]: {
        color: theme.palette.containee.exited,
        backgroundColor: "inherit",
        borderColor: theme.palette.containee.exited,
        borderWidth: 2,
        borderStyle: "solid",
    },

    '&.pod': {
        backgroundColor: theme.palette.containee.pod,
    },
}))

// Groups (Kubernetes) pods with same namespace.
interface PodNamespace {
    name: string
    pods: Pod[]
}

const isPodNamespace = (item: PodNamespace | Containee): item is PodNamespace => (
    (item as PodNamespace).pods !== undefined
)

const sortItemsByName = (itemA: PodNamespace | Containee, itemB: PodNamespace | Containee) => (
    (itemA as Containee).name.localeCompare((itemB as Containee).name)
)

// Given a list of containees, this function tries to find pods with namespaces:
// it then creates dedicated pod namespace items for them and drops the
// namespaced pods from the returned list. These dropped namespaced pods instead
// are referenced through the newly added pod namespace items, so they're not
// lost, but instead in a deeper level of hierarchy.
const namespacify = (containees: Containee[]) => {
    // Determine how certain turtle namespaces of containers map to KinD
    // clusters.
    const kindClusters = kindClusterMap(containees)
    // Skim the containees for pods...
    const podnamespaces: { [key: string]: PodNamespace } = {}
    const filtered = containees.filter(containee => {
        if (isPod(containee) && containee.flavor === PodFlavors.K8SPOD) {
            const [namespace] = podNamespaceAndName(containee)
            const kindCluster = kindClusters[containee.containers[0].turtleNamespace]
            const namespaceOrCluster = (kindCluster && `${kindCluster}: ${namespace}`) || namespace
            let podns = podnamespaces[namespaceOrCluster]
            if (!podns) {
                podns = {
                    name: namespaceOrCluster,
                    pods: []
                } as PodNamespace
                podnamespaces[namespaceOrCluster] = podns
            }
            podns.pods.push(containee)
            // drop pod from containee list, it will be referenced from the pod
            // namespace instead.
            return false
        }
        // non-name-namespaceable containee, so keep it. :D
        return true
    })
    return (Object.values(podnamespaces) as (Containee | PodNamespace)[])
        .concat(filtered)
}


// Returns a map of container turtle namespaces mapping to their corresponding
// KinD cluster names.
const kindClusterMap = (containees: Containee[]) => {
    // Build a map that is keyed by kind cluster containers' names and maps them
    // onto their kind cluster names.
    const kindTurtleNamespaceMap: { [key: string]: string } = {}
    containees
        .filter(containee =>
            isContainer(containee)
            && !!containee.labels['io.x-k8s.kind.cluster'])
        .forEach((container) => {
            kindTurtleNamespaceMap[container.name] = (container as Container).labels['io.x-k8s.kind.cluster']
        })
    return kindTurtleNamespaceMap
}

// Returns the namespace of a pod (if defined) and pod name sans namespace, or
// just the pod name.
const podNamespaceAndName = (pod: Pod, withoutTurtleNamespace: boolean = false): [string, string] => {
    const turtleNamespace = withoutTurtleNamespace ? '' : pod.containers[0].turtleNamespace
    const prefix = turtleNamespace ? `[${turtleNamespace}]:` : ''
    switch (pod.flavor) {
        case PodFlavors.K8SPOD: {
            const parts = pod.name.split('/')
            return parts.length ? [parts[0], prefix + parts.slice(1).join('/')]
                : ['', prefix + parts[0]]
        }
        default:
            return ['', prefix + pod.name]
    }
}


export interface ContaineeNavigatorProps {
    /** all discovered network namespace with their attached containees. */
    allnetns: NetworkNamespaces
    /** 
     * if `true`, the component doesn't render any network namespaces with
     * only loopback interfaces in order to reduce clutter.
     */
    filterEmpty?: boolean
    /** disables links; for help usage only */
    nolink?: boolean
}

/**
 * The `ContaineeNavigator` renders navigation links to all discovered
 * containees, sorted lexicographically (except for the initial process in the
 * initial namespace, the "initsomething").
 *
 * The containee names are adorned by their containee type icons. The icons
 * indicate the containee states by their background color.
 *
 * When the filtering out empty network namespaces is specified, then the
 * navigator will not show any containees of such empty network namespaces.
 */
export const ContaineeNavigator = ({ allnetns, filterEmpty, nolink }: ContaineeNavigatorProps) => {
    const domIdBase = useContextualId('')

    const match1 = useMatch('/:view')
    const match2 = useMatch('/:view/:details')
    const match = (match1 || match2) ? { ...match1, ...match2 } : null

    const inDetails = !!match && !!(match.params as { [key: string]: string })['details']

    // Get all pods and those pesky primitive containees that aren't pot'ed. In
    // case a containee should appear multiple times, it is attached to multiple
    // network namespaces and we're fine with that.
    const allContainees = Object.values(allnetns)
        .filter(netns => !emptyNetns(netns) || !filterEmpty)
        .map(netns =>
            (netns.pods as Containee[]).concat(
                netns.containers.filter(primcontainee => !isPodContainer(primcontainee))))
        .flat() // bang all namespace-local containees into a single flat list.

    // Prepare the item list from the containees where this list now replaces
    // groups of pod'ed containers with the (identifier) namespace with a
    // namespace item. This then allows us to render a more useful two-level
    // hierarchy instead of just flat containees.
    const items = namespacify(allContainees)
        .sort(sortItemsByName)

    const containeeItem = (containee: Containee, dropTurtleNamespace: boolean = false) => {
        const CeeIcon = ContaineeIcon(containee)
        const netns = isPod(containee) ? containee.containers[0].netns : containee.netns
        const view = match && (match.params as { [key: string]: string })['view'] || ''
        const path = match && !nolink ? (inDetails ?
            `/${view}/${encodeURIComponent(containee.name)}` :
            `/${view}#${domIdBase + netnsId(netns)}`)
            : (nolink ? undefined : '.')
        const name = isPod(containee)
            ? podNamespaceAndName(containee, dropTurtleNamespace)[1]
            : containee.name
        const privileged = isPrivilegedContainer(containee)
            ? <PrivilegedIcon />
            : isElevatedContainer(containee) ? <CapableIcon /> : ''

        return <ContaineeItem
            key={`${netns.netnsid}${containee.name}`}
            dense
            to={path || ''}
            component={RouterLink}
            className={clsx(
                (isPrivilegedContainer(containee) && 'privileged') ||
                (isElevatedContainer(containee) && 'elevated'),
            )}
        >
            <ListItemAvatar>
                <StatefulAvatar
                    className={clsx(
                        containeeState(containee),
                        isPod(containee) && 'pod',
                    )}
                ><CeeIcon /></StatefulAvatar>
            </ListItemAvatar>
            {privileged}
            <Typography>{name}</Typography>
        </ContaineeItem>
    }

    return (<>{items.map(item => (
        !isPodNamespace(item)
            ? containeeItem(item)
            : <ListItem
                key={`${item.name}`}
                dense
            >
                <List
                    subheader={<ListSubheader>{`${item.name}/`}</ListSubheader>}
                    dense
                >
                    {item.pods
                        .sort(sortContaineesByName)
                        .map(pod => containeeItem(pod, true))}
                </List>
            </ListItem>
    ))}</>)
}

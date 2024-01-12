// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { Process } from './process'
import { NetworkNamespace, NetworkNamespaces } from './netns'
import { IpAddress } from './address'
import { GHOSTWIRE_LABEL_ROOT } from './model'
import { notDockerDefaultCaps } from 'utils/capabilities'


/**
 * A "Containee" is "something" confined ("contained") in a particular network
 * namespace. It can be either a "primitive" containee, such as a container or a
 * stand-alone process. Or it can be a "pod" where all its grouped containers
 * share the same network namespace.
 */
export type Containee = PrimitiveContainee | Pod

/** Type guard for any containee. */
export const isContainee = (containee: unknown): containee is Containee => (
    !!containee && (
        isPod(containee as Containee)
        || isSandbox(containee as Containee)
        || isBusybox(containee as Containee)
        || isContainer(containee as Containee)
    )
)

/**
 * A "PrimitiveContainee" is an element confined ("contained") by a particular
 * (network) namespace, but not necessarily a container. "Primitive" here is
 * meant as opposed to "grouped" or "composite".
 */
export type PrimitiveContainee = Sandbox | Busybox | Container

/**
 * A "Sandbox" is "something" that is attached to a specific network namespace.
 * In its most basic form, it is just a bind-mount that keeps the referenced
 * network namespace alive. Derived interfaces then bring in further types of
 * "containees", such as Busyboxes with processes, as well as Containers.
 */
export interface Sandbox {
    /** 
     * Name of a "sandbox" which is attached to a specific network namespace.
     */
    name: string
    /** Hierarchy */
    turtleNamespace: string
    /** Network namespace the sandbox is attached to. */
    netns: NetworkNamespace
}

/**
 * Type guard for `Sandbox` containees. Returns true only if the given containee
 * is a Sandbox object, but not anything else, not even a derived type such as a
 * Busybox.
 *
 * @param containee the containee to be Sandbox-type guarded.
 */
export const isSandbox = (containee: Containee): containee is Sandbox => {
    return (containee as Sandbox).netns !== undefined
        && (containee as Busybox).ealdorman === undefined
}

/**
 * A "Busybox" is a "Sandbox", but with at least one process attached to the
 * referenced network namespace. However, a Busybox in itself isn't yet a
 * Container.
 */
export interface Busybox extends Sandbox {
    /**
     * The most senior leader process attached to this Busybox (namespace).
     * Note that there might be no ealdorman in case the container is still
     * undergoing start configuration and its initial container process hasn't
     * been started yet.
     */
    ealdorman: Process
    /** 
     * Set of topmost "leader" processes attached a busy sandbox, according to
     * the process tree. Please note that there might be no leader processes
     * during the initialization phase of a container, before its initial
     * container process starts to run.
     */
    leaders: Process[]
    /** 
     * DNS "client" settings, such as hostname (both from /etc/hostname as well
     * as per the attached UTS namespace), /etc/hosts, and some other things
     * relevant to DNS "clients". And just for the record: officially, DNS
     * clients are actually termed "stub resolvers".
     */
    dns: DNSSettings
}

/**
 * Type guard for `Busybox` containees. Returns true only if the given containee
 * is a Busybox object, but not anything else, not even a derived type such as a
 * Container.
 *
 * @param containee the containee to be Sandbox-type guarded.
 */
export const isBusybox = (containee: Containee): containee is Busybox => {
    return (containee as Busybox).ealdorman !== undefined
        && (containee as Container).engineType === undefined
}

export interface HostAddressBinding {
    name: string
    address: IpAddress
}

export interface DNSSettings {
    utsHostname: string
    etcDomainname: string
    etcHostname: string
    etcHosts: HostAddressBinding[]
    nameservers: IpAddress[]
    searchlist: string[]
}

/**
 * A "Container" is a Busybox managed by some container engine/runtime. And with
 * lots of user-space information required for the user-space "lie" that
 * containers are.
 */
export interface Container extends Busybox {
    /**
     * type of container, such as "docker", "ie-app",... Please note that the
     * flavor might deviate or be further refined from the engine type. For
     * instance, a "docker" engine type container might have the type "ie-app"
     * if it is part of an Industrial Edge application deployment.
     */
    flavor: string
    /** type of container engine/runtime, such as "docker", "containerd", ... */
    engineType: string
    /** 
     * the container's ID as opposed to its name; not every container
     * engine/runtime differentiates between the ID and name of a container; if
     * not, then id and name are the same.
     */
    id: string
    /** container state. */
    state: ContainerState
    /** discovery-related attributes for this container. */
    attrs: { [key: string]: string }
    /** container labels. */
    labels: { [key: string]: string }
    /** grouping pod, if any. */
    pod?: Pod
    /** is it the pod's sandbox? */
    sandbox?: boolean
    /** part of a project? */
    project?: Project
}

/**
 * Type guard for `Container` containees.
 *
 * @param containee the containee to be Sandbox-type guarded.
 */
export const isContainer = (containee: Containee): containee is Container => {
    return containee && (containee as Container).engineType !== undefined
}

/**
 * Returns true if the specified containee is a (Docker) privileged container.
 *
 * @param containee the containee to check for being a privileged twat.
 */
export const isPrivilegedContainer = (containee: Containee): boolean => {
    return isContainer(containee)
        && (GHOSTWIRE_LABEL_ROOT + 'container/privileged' in containee.labels)
}

/**
 * Returns true if the specified containee is a container with "elevated"
 * bounded capabilities beyond the Docker default set of capabilities.
 *
 * @param containee the containee to check for being a privileged twat.
 */
export const isElevatedContainer = (containee: Containee): boolean => {
    return isContainer(containee)
        && !!containee.ealdorman && notDockerDefaultCaps(containee.ealdorman.capbnd)
}

export enum ContaineeTypes {
    /** bind-mounted Sandbox */
    BINDMOUNT = 'bindmount',
    /** stand-alone (un-containerized) process Busybox */
    PROCESS = 'proc', // sic!

    /** Docker engine */
    DOCKER = "docker",
    /** containerd engine */
    CONTAINERD = "containerd",
}

// Not identical to the decorator type and flavor definition, so take care. This
// has some Ghostwire v1 legacy...
export enum ContainerFlavors {
    DOCKER = 'docker',
    DOCKERPLUGIN = 'dockerplugin',
    CONTAINERD = 'containerd',
    IERUNTIME = 'ie-runtime',
    IEAPP = 'ie-app',
    KIND = 'kind',
    PODMAN = 'podman',
    CRI = 'CRI',
}

/**
 * Returns the type of container (engine type), ContaineeTypes.PROCESS, or
 * ContaineeTypes.BINDMOUNT.
 *
 * @param box containee object
 */
export const containeeType = (containee: PrimitiveContainee) => {
    // Type guards to our rescue... ;)
    return (isContainer(containee) && containee.engineType)
        || (isSandbox(containee) && ContaineeTypes.PROCESS)
        || ContaineeTypes.BINDMOUNT
}

/**
 * Sort function ordering containees by their names, yet ensuring that the
 * "init(1)" entity will always be first. Even if it's systemd.
 * 
 * Turtle namespaces are taken into account for sorting order.
 * 
 * @param containeeA a containee object
 * @param containeeB another containee object
 */
export const sortContaineesByName = (containeeA: Containee, containeeB: Containee) => {
    const aIsInit = isBusybox(containeeA) && containeeA.ealdorman.pid === 1
    const bIsInit = isBusybox(containeeB) && containeeB.ealdorman.pid === 1
    if (aIsInit !== bIsInit) {
        return aIsInit ? -1 : 1
    }
    return `${isContainer(containeeA) ? containeeA.turtleNamespace : ''}:${containeeA.name}`
        .localeCompare(`${isContainer(containeeB) ? containeeB.turtleNamespace : ''}:${containeeB.name}`)
}

/**
 * Returns a sufficiently unique and stable key string, to be used for
 * identifying React components.
 * 
 * @param containee containee object
 */
export const containeeKey = (containee: Containee) => {
    return isPod(containee) ? `${containee.name}`
        : `${containee.turtleNamespace}:${containee.name}`
}

export enum ContainerState {
    Exited,
    Running,
    Paused,
    Restarted
}

// containerState converts a textual container state, such as "running" or
// "paused" into its corresponding enumeration value.
export const containerState = (cs: string) => {
    switch (cs) {
        case "exited":
            return ContainerState.Exited
        case "running":
            return ContainerState.Running
        case "pausing":
        case "paused":
            return ContainerState.Paused
        case "restarted":
            return ContainerState.Restarted
    }
    return ContainerState.Running
}

const containerStateStrings = {
    [ContainerState.Exited]: "exited",
    [ContainerState.Running]: "running",
    [ContainerState.Paused]: "paused",
    [ContainerState.Restarted]: "restarted",
}

export const containerStateString = (cs: ContainerState) => containerStateStrings[cs]


// The container type description map only needs entries for container flavors
// which cannot be covered generically.
const containerFlavorDescriptions: { [key: string]: string } = {
    [ContainerFlavors.DOCKERPLUGIN]: 'Managed Docker Plugin',
    [ContainerFlavors.IERUNTIME]: 'Industrial Edge Runtime',
    [ContainerFlavors.IEAPP]: 'Industrial Edge App',
    [ContainerFlavors.KIND]: 'KinD Kuhbernetes node',
    [ContainerFlavors.CRI]: 'k8s CRI-API',
    [ContainerFlavors.PODMAN]: 'Podman',
}

export enum PodFlavors {
    K8SPOD = 'pod',
}

const podFlavorDescriptions: { [key: string]: string } = {
    [PodFlavors.K8SPOD]: 'Kuhbernetes pod',
}

/**
 * Returns a short textual description of a containee based on its type or
 * container flavor, suitable for use in tooltips, et cetera.
 * 
 * @param containee containee object
 */
export const containeeDescription = (containee: Containee) => {
    if (!containee) {
        return '???' // ...always play safe.
    }
    if (isPod(containee)) {
        return podFlavorDescriptions[containee.flavor]
            || `${containee.flavor.charAt(0).toUpperCase}${containee.flavor.slice(1)} pod`
    }
    if (!isContainer(containee)) {
        return isBusybox(containee) ?
            (containee.ealdorman.pid === 1 ?
                'the initial system process with PID 1'
                : 'running stand-alone process')
            : 'bind-mounted network namespace'
    }
    // It's a ... container!
    const flavor = containerFlavorDescriptions[containee.flavor]
        || (containee.flavor.charAt(0) + containee.flavor.slice(1))
    const privileged = isPrivilegedContainer(containee) ? 'privileged ' : ''
    const elevated = !privileged && isElevatedContainer(containee) ? ' with additional non-default capabilities' : ''
    return `${containerStateString(containee.state)} ${privileged}${flavor} container${elevated}`
}

/**
 * Returns the state of a container, or '' if the given containee is a
 * bind-mount. Thus, stand-alone processes as well as pods will always be
 * considered to be in the running state.
 *
 * @param containee containee object
 */
export const containeeState = (containee: Containee) => {
    if (isContainer(containee)) {
        return containerStateString(containee.state)
    }
    return isBusybox(containee) ?
        containerStateString(ContainerState.Running) : ''
}

/**
 * Returns the "display name" of a containee: this is the name of a sandbox or
 * busybox, but in case of containers this isn't always the container name as
 * reported by the responsible container engine. For CRI-managed containers this
 * is the name as set via CRI, not the container ID (name).
 *
 * @param containee containee object
 */
export const containeeDisplayName = (containee: Containee) => {
    if (!isContainer(containee)) {
        return containee.name
    }
    return containee.labels['io.kubernetes.container.name'] || containee.name
}

/**
 * A "Pod" is a group of containers, all sharing the same network namespace. The
 * term "pod" here refers to the general concept, not a particular flavor such
 * as Kubernetes pods.
 *
 * **Note:** nothing prevents multiple pods sharing the same network namespace,
 * as can be seen from Kubernetes-in-Docker ("KinD").
 */
export interface Pod {
    /** name of pod. */
    name: string
    /** pod flavor. */
    flavor: string
    /** tightly grouped containers. */
    containers: Container[]
}

export const isPod = (containee: Containee): containee is Pod => (
    (containee as Pod).containers !== undefined
)

/**
 * Returns true if containee is a container that is part of a Kubernetes pod,
 * false otherwise.
 *
 * @param containee containee object
 * @returns true if containee is a container that is part of a Kubernetes pod.
 */
export const isPodContainer = (containee: Containee) => (
    isContainer(containee) && containee.pod !== undefined
)

/**
 * A "Project" is a group of containers all belonging the same (Docker) composer
 * group. Please note that
 */
export interface Project {
    /** project name. */
    name: string
    /** project flavor. */
    flavor: string
    /** 
     * losely "composed" containers, which might even be in other shared network
     * namespaces.
     */
    containers: Container[]
    /** network namespaces used by the composed containers. */
    netnses: NetworkNamespaces
}

export type NetworkNamespaceOrProject = NetworkNamespace | Project

/**
 * Returns true if project is a project (and not a network namespace).
 * 
 * @param project project object.
 * @returns true if project is a project.
 */
export const isProject = (project: NetworkNamespaceOrProject): project is Project => (
    // didn't we told you that namespaces do not have names???
    (project as Project).name !== undefined
)

export enum ProjectFlavors {
    COMPOSER = ContainerFlavors.DOCKER,
    IEAPP = ContainerFlavors.IEAPP,
}

const projectFlavorDescriptions: { [key: string]: string } = {
    [ProjectFlavors.COMPOSER]: 'Docker composer project',
    [ProjectFlavors.IEAPP]: 'Industrial Edge app project',
}

/**
 * Returns a description of the type of project specified, or an empty
 * description if not known.
 * 
 * @param project project object.
 * @returns descriptive text.
 */
export const projectDescription = (project: Project) => {
    if (!project) {
        return '???'
    }
    return projectFlavorDescriptions[project.flavor] || `${project.flavor} project`
}

/**
 * Returns the name of the (Docker) composer project the containee belongs to,
 * if any, or "".
 * 
 * @param containee containee object
 * @returns name of (Docker) composer project, if any, or "".
 */
export const inProject = (containee: Containee) => (
    (containee && isContainer(containee)
        && containee.labels && containee.labels['com.docker.compose.project'])
    || ""
)

/**
 * Returns allmost all containees, except for containers that are part of pods.
 *
 * @param netns network namespace object
 */
export const containeesOfNetns = (netns: NetworkNamespace) => (
    (netns.pods as Containee[])
        .concat(netns.containers
            .filter(containee => !isPodContainer(containee)))
)

export const containeeFullName = (containee: Containee) => {
    // Determine the containee name, taking an optional turtle namespace into
    // account. But as pods do not have a turtle namespace, but only their
    // pod'ed containers, we then go for the first container in a pod.
    const turtleNamespace =
        (isPod(containee) && containee.containers[0].turtleNamespace)
        || (isContainer(containee) && containee.turtleNamespace)
    return turtleNamespace ? `[${turtleNamespace}]:${containee.name}` : containee.name
}

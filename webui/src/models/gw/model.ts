// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import JSBI from 'jsbi'

import { PrimitiveContainee, Busybox, Container, containerState, HostAddressBinding, ContaineeTypes, ContainerFlavors, Pod, isContainer, Project, NetworkNamespaceOrProject } from './containee'
import { NetworkInterface, SRIOVRole, TapTunProcessor } from './nif'
import { Process } from './process'
import { AddressFamily, addressFamilyByName, IpAddress } from './address'
import { IpRoute } from './route'
import { NetworkNamespace, NetworkNamespaces } from './netns'
import { PortUser, TransportPort } from './ports'
import { isBusybox } from '.'
import { ForwardedPort } from './forwardedports'

/* Ghostwire's engine v2 own label namespace for passing additional information. */
export const GHOSTWIRE_LABEL_ROOT = 'gostwire/'
export const TURTLEFINDER_LABEL_ROOT = "turtlefinder/"

export const hiddenLabel = (key: string) =>
    [
        GHOSTWIRE_LABEL_ROOT,
        TURTLEFINDER_LABEL_ROOT,
        "github.com/thediveo/whalewatcher/",
        "github.com/thediveo/lxkns/",
    ].find(root => key.startsWith(root)) !== undefined

/**
 * Ghostwire "full" discovery API results
 */
export interface Discovery {
    networkNamespaces: NetworkNamespaces
    hostname?: string
    kubernetesNode?: string
    metadata: { [key: string]: any }
}

export const fromjson = (jsondata) => {
    const disco = {
        networkNamespaces: {},
        metadata: jsondata.metadata,
    } as Discovery

    // Convert the v1 container representation into the future v2
    // representation where containers and processes are separate, in line
    // with the Gostwire and lxkns models.
    //
    // We also build here a map of the JSON document identifiers seen for the
    // network interfaces, in order to later be able to resolve references.
    // The same also applies to containees ("containers" in Ghostwire v1
    // parlance).
    const nifmap: { [key: string]: NetworkInterface } = {}
    const containeemap: { [key: string]: PrimitiveContainee } = {}
    const podmap: { [key: string]: Pod } = {}
    const projectmap: { [key: string]: Project } = {}

    jsondata['network-namespaces'].forEach(jnetns => {
        const newnetns = {
            netnsid: jnetns.netnsid,
            isInitial: false,
        } as NetworkNamespace
        disco.networkNamespaces[newnetns.netnsid] = newnetns

        // Process the list of pods (namespace-tightly grouped containers)
        newnetns.pods = jnetns['container-groups'].map(jpod => {
            const pod: Pod = {
                name: jpod.name,
                flavor: jpod.type,
                containers: [],
            }
            podmap[jpod.id] = pod
            return pod
        })

        // Process the list of containers attached to this network namespace
        // and separate out their processes.
        newnetns.containers = jnetns.containers
            .filter(jcntr =>
                // https://github.com/siemens/ghostwire/issues/5
                !(jcntr.name.startsWith('containerd-shim')
                    && jcntr.cmdline && jcntr.cmdline.includes(" -namespace moby ")))
            .map(jcntr => {
                const primitiveBox: PrimitiveContainee = {
                    name: jcntr.name,
                    turtleNamespace: "",
                    netns: newnetns,
                }
                containeemap[jcntr.id] = primitiveBox
                if (jcntr.type !== ContaineeTypes.BINDMOUNT) {
                    // It's not a bindmount, so at least something with one or
                    // more processes...
                    const bbox = primitiveBox as Busybox
                    const proc: Process = {
                        pid: jcntr.pid,
                        pidnsid: jcntr.pidns,
                        capbnd: JSBI.BigInt(jcntr.capbnd || 0), // safety fallback
                        name: jcntr.name,
                        cmdline: jcntr.cmdline.split(' '),
                    }
                    bbox.ealdorman = proc
                    bbox.leaders = [proc]
                    // Because it has process(es), we also have its DNS
                    // configuration; as this is an application-layer
                    // configuration, it cannot exist outside a process, but
                    // instead in the mounted filesystem(s) seen by a process.
                    bbox.dns = {
                        etcHostname: jcntr.dns['etc-hostname'] || "",
                        etcDomainname: jcntr.dns.domainname || "",
                        utsHostname: jcntr.dns['uts-hostname'] || "",
                        etcHosts: jcntr.dns['etc-hosts'].map(host => {
                            const family = host.address.includes(':') ? AddressFamily.IPv6 : AddressFamily.IPv4
                            return {
                                name: host.name,
                                address: {
                                    address: host.address,
                                    family: family,
                                    prefixlen: family === AddressFamily.IPv6 ? 128 : 32,
                                } as IpAddress,
                            } as HostAddressBinding
                        }),
                        nameservers: (jcntr.dns.nameservers || []).map(addr => {
                            const family = addr.includes(':') ? AddressFamily.IPv6 : AddressFamily.IPv4
                            return {
                                address: addr,
                                family: family,
                                prefixlen: family === AddressFamily.IPv6 ? 128 : 32,
                            } as IpAddress
                        }),
                        searchlist: jcntr.dns.searchlist ? [jcntr.dns.searchlist].flat() : [],
                    }
                    // Did we stumble upon the initial network namespace? Then
                    // mark it!
                    if (proc.pid === 1) {
                        newnetns.isInitial = true
                    }
                }
                if (jcntr.type !== ContaineeTypes.BINDMOUNT
                    && jcntr.type !== ContaineeTypes.PROCESS) {
                    // it's a container thingy...
                    const cbox = primitiveBox as Container
                    cbox.labels = jcntr.labels || {}
                    cbox.engineType = jcntr.type
                    // For the container flavor, assume it's the same as the
                    // (engine) type for now.
                    cbox.flavor = cbox.engineType
                    // If the discovery engine gave us a container ID (not to be
                    // confused with the document container ID, in "id"), then take
                    // that, otherwise fall back to the container name.
                    cbox.id = jcntr['container-id'] || jcntr.name
                    cbox.state = containerState(jcntr.status)
                    cbox.turtleNamespace = jcntr.prefix || ""
                    // is it pod'ed? Then resolve the reference and birectionally
                    // relate this container with its pod.
                    if (jcntr.group) {
                        cbox.pod = podmap[jcntr.group]
                        cbox.pod.containers.push(cbox)
                    }
                    if (jcntr.sandbox !== undefined) {
                        cbox.sandbox = !!jcntr.sandbox // make sure it's boolean
                    }
                    // is it part of a (Docker) composer project?
                    cbox.project = null
                    if (jcntr.labels && !!jcntr.labels['com.docker.compose.project']) {
                        const projectname = jcntr.labels['com.docker.compose.project']
                        var project = projectmap[projectname]
                        if (project === undefined) {
                            project = {
                                name: projectname,
                                flavor: cbox.flavor,
                                containers: [],
                                netnses: {},
                            } as Project
                            projectmap[projectname] = project
                        }
                        cbox.project = project
                        project.containers.push(cbox)
                        project.netnses[cbox.netns.netnsid] = cbox.netns
                    }
                }
                return primitiveBox
            })
        // Process the list of network interfaces belonging to this network
        // namespace. Also add the network interfaces to the map, index by
        // their unique JSON document identifiers.
        newnetns.nifs = jnetns['network-interfaces'].map(jnif => {
            const nif: NetworkInterface = {
                netns: newnetns,
                name: jnif.name,
                alias: jnif.alias || "",
                index: jnif.index,
                kind: jnif.kind,
                operstate: jnif.operstate,
                isPhysical: jnif.physical,
                isPromiscuous: jnif.promisc,
                sriovrole: jnif['sr-iov-role'] || SRIOVRole.None,
                addresses: undefined,
                labels: jnif.labels || {},
            }
            if (jnif.tuntap) {
                nif.tuntapDetails = {
                    mode: jnif.tuntap.mode,
                    processors: [],
                }
            }
            if (jnif.vxlan) {
                nif.vxlanDetails = {
                    vid: jnif.vxlan.vid,
                    arpProxy: jnif.vxlan['arp-proxy'],
                    remotePort: jnif.vxlan['remote-port'],
                    sourcePortLow: jnif.vxlan['source-portrange'].low,
                    sourcePortHigh: jnif.vxlan['source-portrange'].high,
                }
            }
            if (jnif.vlan) {
                nif.vlanDetails = {
                    vid: jnif.vlan.vid,
                    vlanProtocol: jnif.vlan['vlan-protocol']
                }
            }
            nifmap[jnif.id] = nif

            // Read in all network addresses assigned to this interface...
            nif.addresses = ['ipv4', 'ipv6'].map(addrfamily =>
                jnif.addresses[addrfamily].map(addr => ({
                    address: addr.address,
                    prefixlen: addr.prefixlen,
                    family: addr.family,
                    preferredLifetime: addr['preferred-lifetime'],
                    validLifetime: addr['valid-lifetime'],
                    scope: addr.scope,
                } as IpAddress))
            ).flat()
            nif.addresses.push({
                address: jnif.addresses.mac,
                prefixlen: 0,
                family: AddressFamily.MAC,
            } as IpAddress)

            return nif
        })
        // Read in the routes in this network namespace...
        newnetns.routes = ['ipv4', 'ipv6'].map(addrfamily =>
            jnetns.routes[addrfamily].map(route => ({
                destination: route.destination,
                prefixlen: route['destination-prefixlen'],
                family: route.family,
                index: route.index,
                nif: nifmap[route['network-interface-idref']],
                nexthop: route['next-hop'],
                preference: route.preference,
                priority: route.priority,
                table: route.table,
                type: route.type,
            } as IpRoute))
        ).flat()
    })
    // Resolve the references between network interfaces ... now these
    // references are one of the truely unique features of Ghostwire!
    Object.values(jsondata['network-namespaces']).forEach(jnetns =>
        jnetns['network-interfaces'].forEach(jnif => {
            const nif = nifmap[jnif.id]
            if (jnif.macvlans) {
                nif.macvlans = jnif.macvlans.map(nif => nifmap[nif.idref])
            }
            if (jnif.slaves) {
                // Nota bene: while the discovery service classifies VXLAN
                // overlays as slaves, we now sort it out; instead, we are
                // maintaining a dedicated overlay list.
                nif.slaves = jnif.slaves.map(nif => nifmap[nif.idref])
                    .filter(slave => !slave.vxlanDetails)
            }
            jnif.pf && (nif.pf = nifmap[jnif.pf.idref])
            jnif.master && (nif.master = nifmap[jnif.master.idref])
            jnif.macvlan && (nif.macvlan = nifmap[jnif.macvlan.idref])
            jnif.peer && (nif.peer = nifmap[jnif.peer['peer-idref']])
            if (jnif.vxlan) {
                // this VXLAN is the overlay, resolve our underlay reference,
                // and then backlink the underlay to us (the overlay).
                nif.underlay = nifmap[jnif.vxlan.idref]
                if (!nif.underlay.overlays) {
                    nif.underlay.overlays = []
                }
                nif.underlay.overlays.push(nif)
            }
            // TAP/TUNs don't reference other network interfaces, but processes
            // ... but hey. we need to resolve this relation, too!
            if (jnif.tuntap) {
                nif.tuntapDetails.processors = jnif.tuntap.processors.map(
                    proc => ({
                        cmdline: proc.cmdline.split(' '),
                        containee: containeemap[proc['container-idref']],
                        pid: proc.pid,
                    } as TapTunProcessor))
            }
        }))
    // Only now read in the transport port related discovery information,
    // because now we can easily resolve the references from port users to
    // their containees.
    Object.values(jsondata['network-namespaces']).forEach((jnetns: any) => {
        const netns = disco.networkNamespaces[jnetns.netnsid]
        // Read in the open transport-layer ports in this network namespace...
        netns.transportPorts = ['ipv4', 'ipv6'].map(addrfamily => {
            const family = addressFamilyByName(addrfamily)
            const prefixlen = family === AddressFamily.IPv6 ? 128 : 32
            return jnetns['transport-ports'][addrfamily].map(port => ({
                state: port.state,
                macroState: port.macrostate,
                protocol: port.protocol,
                localAddress: {
                    address: port['local-address'],
                    family: family,
                    prefixlen: prefixlen,
                } as IpAddress,
                localPort: port['local-port'],
                localServicename: port['local-servicename'],
                remoteAddress: {
                    address: port['remote-address'],
                    family: family,
                    prefixlen: prefixlen,
                } as IpAddress,
                remotePort: port['remote-port'],
                remoteServicename: port['remote-servicename'],
                v4mapped: port.v4mapped,
                users: port.owners.map(owner => ({
                    cmdline: owner.cmdline.split(' '),
                    containee: containeemap[owner['container-idref']],
                    pid: owner.pid,
                } as PortUser)),
            } as TransportPort))
        }
        ).flat()
    })
    // Then read in the forwarded port related discovery information, so that we
    // can easily resolve the references from port users to their containees.
    Object.values(jsondata['network-namespaces']).forEach((jnetns: any) => {
        const netns = disco.networkNamespaces[jnetns.netnsid]
        netns.forwardedPorts = ['ipv4', 'ipv6'].map(addrfamily => {
            const family = addressFamilyByName(addrfamily)
            const prefixlen = family === AddressFamily.IPv6 ? 128 : 32
            return jnetns['forwarded-ports'][addrfamily].map(port => {
                const fwaddr = port['forward-ip']
                const fwfam = fwaddr.includes(':') ? AddressFamily.IPv6 : AddressFamily.IPv4
                const fwprefixlen = fwfam === AddressFamily.IPv6 ? 128 : 32
                return {
                    protocol: port.protocol,
                    address: {
                        address: port['ip'],
                        family: family,
                        prefixlen: prefixlen,
                    } as IpAddress,
                    port: port['port'],
                    servicename: port['servicename'],
                    forwardedAddress: {
                        address: fwaddr,
                        family: fwfam,
                        prefixlen: fwprefixlen,
                    } as IpAddress,
                    forwardedPort: port['forward-port'],
                    forwardedServicename: port['forward-servicename'],
                    netns: disco.networkNamespaces[port['netnsid']],
                    users: port.owners.map(owner => ({
                        cmdline: owner.cmdline.split(' '),
                        containee: containeemap[owner['container-idref']],
                        pid: owner.pid,
                    } as PortUser)),
                } as ForwardedPort
            })
        }
        ).flat()
    })
    // And now for something completely different: sniff for the presence of an
    // Industrial Edge runtime and applications...
    const allContainers = Object.values(disco.networkNamespaces)
        .map(netns => netns.containers)
        .flat()
        .filter(containee => isContainer(containee)) as Container[]
    const ieRuntime = allContainers.find(
        cntr => cntr.name === 'edge-iot-core' && cntr.engineType === ContaineeTypes.DOCKER)
    if (ieRuntime) {
        ieRuntime.flavor = ContainerFlavors.IERUNTIME
        allContainers.forEach(cntr => {
            if (cntr.engineType === 'docker'
                && Object.keys(cntr.labels).find(key => key.startsWith('com_mwp_conf_'))) {
                // It's an IE app container...
                cntr.flavor = ContainerFlavors.IEAPP
            }
        })
    }
    // And now try to detect KinD containers...
    allContainers.forEach(cntr => {
        if (cntr.labels['io.x-k8s.kind.cluster']) {
            cntr.flavor = ContainerFlavors.KIND
        }
    })
    // Set the correct flavor for managed Docker plugins -- as we're still using
    // the Ghostwire v1 REST API and that doesn't differentiate between
    // container type and container flavor.
    allContainers.forEach(cntr => {
        if (cntr.labels['com.docker/engine.bundle.path']) {
            cntr.flavor = ContainerFlavors.DOCKERPLUGIN
        }
    })
    // Associate composer projects with network namespaces, where possible. This
    // is later used to quickly detect when to group network namespaces
    // belonging to the same project.
    //
    // Now, there's a catch here, turned up from under its carpet only by the
    // underlying new lxkns-based v2 Gostwire discovery engine. The lxkns
    // namespace discovery engine doesn't anymore cut some corners as did the
    // braindead Ghostwire v1 discovery engine and thus doesn't make any false
    // v1 assumptions anymore.
    //
    // Especially when Packetflix attaches a "stray" dumpcap process to another
    // network namespace, and container in a composer project in particular,
    // lxkns now clearly shows that there is another containee present, but not
    // from the same composer project. We thus accept stand-alone processes as
    // additional leader proesses in the network namespace of a composer
    // project-originating container -- as long as it's not PID 1. PID 1 would
    // mean that we're in the initial network namespace and thus any container
    // attached to it will still loose its visual project grouping.
    Object.values(disco.networkNamespaces).forEach(netns => {
        // Do all containers in this network namespace share the exactly same
        // project? Here, undefined means we never saw any project, but null
        // means "oh we saw multiple projects or something else".
        const project = netns.containers
            // Sort containers to the front, as this simplifies the following
            // check of determining all containees to be acceptable composer
            // project inhabitants or not, including stray leaders processes.
            .sort((cntrA, cntrB) => {
                const isCntrA = isContainer(cntrA)
                const isCntrB = isContainer(cntrB)
                if (isCntrA || isCntrB) {
                    return (+isCntrB) - (+isCntrA) // https://stackoverflow.com/a/7820695
                }
                return 0 // both aren't containers, so we don't care.
            })
            .reduce((project, cntr) => {
                if (isContainer(cntr)) {
                    if (project === undefined) {
                        return cntr.project // either associated project or null
                    }
                    return cntr.project === project ? project : null
                }
                // It's a stand-alone process: we tolerate them as guests in a
                // project as long as they're not PID 1...
                if (!!project && isBusybox(cntr) && cntr.ealdorman.pid !== 1) {
                    return project
                }
                return null
            },
                undefined as Project) || null
        netns.project = project
    })
    // Fix composer project flavors; projects can only exist from our point of
    // view when there is at least a single, lone container in a project.
    Object.values(projectmap).forEach(project => { project.flavor = project.containers[0].flavor })

    return disco
}

/**
 * Sort order compare function which orders namespaces based on the names of
 * their "containees" (that is, names of containers and stand-alone processes
 * attached to these namespaces).
 *
 * There is one exception, however: the initial network namespace will always be
 * ordered first.
 *
 * Oh, and another exception: project names will also be taken into account and
 * basically work as containee name prefixes ... unless the containee is guest
 * to another network namespace.
 *
 * @param netnsA one network namespace
 * @param netnsB another network namespace
 */
export const orderNetnsByContainees = (netnsA: NetworkNamespace, netnsB: NetworkNamespace) => {
    if (netnsA.isInitial !== netnsB.isInitial) {
        return netnsA.isInitial ? -1 : 1
    }
    // bang together all containee names from the first network namespace,
    // sorted lexicographically and separated by dashes.
    const a = netnsA.containers.map(cntr => cntr.name)
        .sort((n1, n2) => n1.localeCompare(n2))
        .join('-')
    // if all containees come from the same project then prepend that project
    // name, otherwise just take the banged-together containee names.
    const fqa = netnsA.project
        ? netnsA.project.name + "-" + a
        : a
    // same for second network namespace.
    const b = netnsB.containers.map(cntr => cntr.name)
        .sort((n1, n2) => n1.localeCompare(n2))
        .join('-')
    const fqb = netnsB.project
        ? netnsB.project.name + "-" + b
        : b
    return fqa.localeCompare(fqb)
}

/**
 * Returns the sorted list of (1) non-composer group-related network namespaces
 * as well as (2) docker composer groups. So, instead of many individual network
 * namespaces for container belonging to the same project, only a single docker
 * composer group will be returned instead.
 *
 * @param netnses list of all discovered network namespaces.
 */
export const sortedNetnsProjects = (netnses: NetworkNamespace[]) => {

    const list: NetworkNamespaceOrProject[] = []
    const projects: Project[] = [] // list of projects already seen

    netnses.sort(orderNetnsByContainees)
        .forEach(netns => {
            // If its a network namespace without a project, then always add it
            // to the result list; if it has an associated project then don't
            // add the network namespace directly, but only the project, and
            // that only once.
            if (!netns.project) {
                list.push(netns)
            } else {
                if (!projects.includes(netns.project)) {
                    list.push(netns.project)
                    projects.unshift(netns.project)
                }
            }
        })

    return list
}

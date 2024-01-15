// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { Container, GHOSTWIRE_LABEL_ROOT, NetworkInterface, isContainer } from "models/gw"

// Service represents a ("neighborhood") service reachable on a set of networks,
// and with the individual containers powering the service.
export interface Service {
    name: string /** name of (Docker composer) service */
    containers: Container[] /** individual containers powering this service */
    networks: string[] /** names of networks on which this service is available */
}

// Given a container, find the Docker networks (in form of bridge network
// interfaces) it is attached to.
const attachedNetworks = (cntr: Container) => {
    if (cntr.engineType !== 'docker') {
        return []
    }
    const bridges = Object.values(cntr.netns.nifs)
        .filter(nif => nif.kind === 'veth'
            && nif.peer && nif.peer.master
            && !!nif.peer.master.labels[GHOSTWIRE_LABEL_ROOT + 'network/name'])
        .map(nif => nif.peer?.master)
        .filter((bridge, pos, arr) => arr.indexOf(bridge) === pos) as NetworkInterface[]
    const networks = bridges.filter(br => br && !!br.labels[GHOSTWIRE_LABEL_ROOT + 'network/name'])
    return networks
}

// Given a specific ("target") container and a deduplicated list of services,
// return the names of the networks this container is attached to. Returns an
// empty array of network names in case of the target not being part of any
// service.
const networksOfContainer = (target: Container, services: Service[]) => {
    const service = services.find(service => service.containers.includes(target))
    return service ? service.networks : []
}

// returns true if a and b share at least one container.
export const shareContainers = (a: Container[], b: Container[]) =>
    a.find(cntr => b.includes(cntr)) !== undefined

//  merges containers from a and b without duplicates.
export const mergeContainers = (a: Container[], b: Container[]) =>
    [...a, ...b.filter(cntr => !a.includes(cntr))]

// deduplicate a list of services, based on the service names *AND* the
// containers powering the service. The service names in themselves might not
// necessarily "globally" unique within a Docker host. Stand-alone containers
// that are not part of a service but still connected to a non-default network
// are only de-duplicated with respect to the networks the container is attached
// to.
const deduplicate = (services: Service[]) => {
    // index by service name; if no service name defined, then use a pseudo
    // service name derived from s stand-alone container's name.
    const uniques: { [key: string]: Service } = {}
    services.forEach(service => {
        const servicename = service.name || `:${service.containers[0].name}`
        const userv = uniques[servicename]
        if (!userv) {
            uniques[servicename] = service
        } else {
            uniques[servicename].containers = mergeContainers(uniques[servicename].containers, service.containers)
            const uniquenetworks = uniques[servicename].networks
            uniques[servicename].networks = uniquenetworks.concat(
                service.networks.filter((item) => uniquenetworks.indexOf(item) < 0))
        }
    })
    return Object.values(uniques)
}

// returns the Service[]s in the "neighborhood" of the start container: that is,
// services with containers attached to networks that are also attached to the
// start container.
export const neighborhoodServices = (start: Container) => {
    if (!start) {
        return []
    }
    const services = attachedNetworks(start)
        .filter(net => net?.labels[GHOSTWIRE_LABEL_ROOT + 'network/default-bridge'] !== 'true') // skip Docker's default network/bridge
        .map(net => {
            const networkname = net.labels[GHOSTWIRE_LABEL_ROOT + 'network/name']
            const services: { [key: string]: Service } = {}
            net.slaves
                ?.filter(nif => nif.kind === 'veth' && nif.peer)
                .map(nif => nif.peer?.netns.containers.filter(cntr => isContainer(cntr)) as Container[])
                .reduce((cntrs, morecntrs) => mergeContainers(cntrs, morecntrs), [])
                .forEach(cntr => {
                    // Nota bene: consider service-less containers to be
                    // empty-named services with exactly a single container for
                    // it. Also do not lump service-less containers together,
                    // but instead keep them as separate pseudo-services without
                    // service name.
                    const servicename = cntr.labels['com.docker.compose.service'] || `:${cntr.name}`
                    const service = services[servicename]
                    if (service) {
                        service.containers.push(cntr)
                        return
                    }
                    services[servicename] = {
                        name: cntr.labels['com.docker.compose.service'] || '',
                        containers: [cntr],
                        networks: [networkname],
                    } as Service
                })
            return Object.values(services)
        })
        .flat()
    const dedups = deduplicate(services)
    // remove networks which aren't reachable from the "start" container but
    // turned up as we discovered other services and containers connected to
    // multiple networks, but not necessarily all the same as the "start"
    // container is connected to.
    const startNetworks = networksOfContainer(start, dedups)
    dedups.forEach(service => {
        service.networks = service.networks.filter(netw => startNetworks.includes(netw))
    })
    return dedups
}

export const neighborhoodServiceContainerCount = (services: Service[]) =>
    services.reduce((total, service) => total + service.containers.length, 0)

export const sortServices = (a: Service, b: Service) => {
    const snameA = a.name
    const snameB = b.name
    const verdict = snameA.localeCompare(snameB)
    if (verdict !== 0 || snameA !== '' ) {
        return verdict
    }
    // name-less pseudo service entries for a single stand-alone container. We
    // know this since we only arrive here then both services are name-less.
    return a.containers[0].name.localeCompare(b.containers[0].name)
}

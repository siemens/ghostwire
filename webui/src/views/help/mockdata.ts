import JSBI from 'jsbi'
import { AddressFamily, Busybox, Container, ContainerState, HostAddressBinding, IpAddress, NetworkInterface, NetworkNamespace, OperationalState, Pod, Process, Sandbox, SRIOVRole } from 'models/gw'

// Mock data needed here solely for illustrational purposes, to explain
// Ghostwire's display to unsuspecting users.
export const standaloneBox: Busybox = {
    name: 'init (1)',
    turtleNamespace: '',
    netns: null,
    ealdorman: {} as Process,
    leaders: [],
    dns: null,
}

export const containerBoxProcess: Process = {
    cmdline: ['/bin/fool', '--worlddomination'],
    name: 'fool',
    pid: 12345,
    pidnsid: 40961234,
    capbnd: JSBI.BigInt(0),
}

export const containerBox: Container = {
    name: 'Morbid_Moby_1',
    turtleNamespace: '',
    netns: null,
    ealdorman: containerBoxProcess,
    leaders: [containerBoxProcess],
    flavor: 'docker',
    engineType: 'docker',
    id: '1234567890',
    state: ContainerState.Running,
    attrs: {},
    labels: {
        foo: 'bar',
        bar: 'baz',
    },
    dns: {
        etcDomainname: 'edge.venture',
        etcHostname: 'morbidmoby',
        utsHostname: 'morbidmoby',
        etcHosts: [
            {name: 'moby', address: {
                family: AddressFamily.IPv4,
                address: '8.7.6.5',
                prefixlen: 32,
            }} as HostAddressBinding,
        ],
        nameservers: [{
            family: AddressFamily.IPv4,
            address: '8.8.8.8',
            prefixlen: 32,
        } as IpAddress, {
            family: AddressFamily.IPv6,
            address: '2001:db8:f00d:cafe::b10c:e',
            prefixlen: 128,
        } as IpAddress],
        searchlist: ['here.org', 'there.com'],
    },
}

export const boundBox: Sandbox = {
    name: '1234567890',
    turtleNamespace: '',
    netns: null,
}

export const initNetns: NetworkNamespace = {
    containers: [standaloneBox],
    isInitial: true,
    netnsid: 42,
    nifs: {},
    routes: [],
    transportPorts: [],
    forwardedPorts: [],
    pods: [],
}

standaloneBox.netns = initNetns

export const morbidNetns: NetworkNamespace = {
    containers: [containerBox],
    isInitial: true,
    netnsid: 12345678,
    nifs: {},
    routes: [],
    transportPorts: [
        /*
        {
            state: 'listening',
        } as TransportPort,
        */
    ],
    forwardedPorts: [],
    pods: [],
}

containerBox.netns = morbidNetns

export const veth1Nif: NetworkInterface = {
    name: 'veth123456',
    netns: initNetns,
    index: 1,
    kind: 'veth',
    operstate: OperationalState.Up,
    isPhysical: false,
    isPromiscuous: false,
    sriovrole: SRIOVRole.None,
    addresses: [{
        family: AddressFamily.IPv4,
        address: '1.2.3.4',
        prefixlen: 32,
    } as IpAddress],
    peer: undefined,
    labels: {},
}

initNetns.nifs['1'] = veth1Nif

export const veth2Nif: NetworkInterface = {
    name: 'eth0',
    netns: morbidNetns,
    index: 1,
    kind: 'veth',
    operstate: OperationalState.Up,
    isPhysical: false,
    isPromiscuous: false,
    sriovrole: SRIOVRole.None,
    addresses: [{
        family: AddressFamily.IPv4,
        address: '8.7.6.5',
        prefixlen: 32,
    } as IpAddress],
    peer: veth1Nif,
    labels: {},
}

veth1Nif.peer = veth2Nif
morbidNetns.nifs['1'] = veth2Nif

export const podBox: Container = {
    ...containerBox,
    name: 'k8s_zordid_moby_mm_1',
}

export const pod: Pod = {
    name: 'zordid/moby',
    flavor: 'pod',
    containers: [podBox],
}

podBox.pod = pod

export const eth0: NetworkInterface = {
    name: 'eth0',
    netns: initNetns,
    index: 666,
    kind: '',
    operstate: OperationalState.Up,
    isPhysical: true,
    isPromiscuous: false,
    sriovrole: SRIOVRole.None,
    addresses: [{
        family: AddressFamily.IPv4,
        address: '1.2.3.4',
        prefixlen: 32,
    } as IpAddress],
    labels: {},
}

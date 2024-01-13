// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import NetworkNamespaceIcon from 'icons/containees/Netns'
import ContainerIcon from 'icons/containees/Container'
import DockerIcon from 'icons/containees/Docker'
import ContainerdIcon from 'icons/containees/Containerd'
import IERuntimeIcon from 'icons/containees/IERuntime'
import IEAppIcon from 'icons/containees/IEApp'
import KinDIcon from 'icons/containees/Kind'

import PodIcon from 'icons/containees/Pod'
import K8sPodIcon from 'icons/containees/K8sPod'

import { isContainer, isPod, Containee, ContainerFlavors, PodFlavors, isBusybox } from 'models/gw'
import { IEAppProjectIcon } from './appicon'
import DockerManagedPluginIcon from 'icons/containees/DockerManagedPlugin'
import InitialIcon from 'icons/containees/Initial'
import PodmanIcon from 'icons/containees/Podman'
import { SvgIcon, SvgIconProps } from '@mui/material'
import CRIIcon from 'icons/containees/CRI'
import TerminalIcon from '@mui/icons-material/Terminal'
import CaptureIcon from 'icons/Capture'


const ContaineeTypeIcons: { [key: string]: (props: SvgIconProps) => JSX.Element } = {
    'netns': NetworkNamespaceIcon,
    'initialnetns': InitialIcon,
    'unknowntype': ContainerIcon,
    [ContainerFlavors.DOCKER]: DockerIcon,
    [ContainerFlavors.DOCKERPLUGIN]: DockerManagedPluginIcon,
    [ContainerFlavors.CONTAINERD]: ContainerdIcon,
    [ContainerFlavors.IERUNTIME]: IERuntimeIcon,
    [ContainerFlavors.IEAPP]: IEAppIcon,
    [ContainerFlavors.KIND]: KinDIcon,
    [ContainerFlavors.PODMAN]: PodmanIcon,
    [ContainerFlavors.CRI]: CRIIcon,
}

const PodTypeIcons: { [key: string]: (props: SvgIconProps) => JSX.Element } = {
    [PodFlavors.K8SPOD]: K8sPodIcon,
}

/**
 * Returns a containee type icon (more precise: a type icon constructor) based
 * on the type and flavor of containee specified.
 *
 * @param containee containee object, such as a container, stand-alone process,
 * or bind-mount.
 * @param forceStdIcon force to use generic IE runtime/app icons instead of IE
 * app-specific icons when available.
 */
export const ContaineeIcon = (containee: Containee, forceStdIcon?: boolean) => {
    if (isPod(containee)) {
        return PodTypeIcons[containee.flavor] || PodIcon
    }
    if (!isContainer(containee)) {
        return (containee.netns && containee.netns.isInitial)
            ? ContaineeTypeIcons.initialnetns
            : netnsIcon(containee)
    }
    // Now try to find a suitable container-flavor icon, or fall back to our
    // generic one. Please note that the type checker has correctly noticed us
    // using the Congtainer type guard above and has concluded that at this
    // point in the code the containee variable can only be of interface
    // Container ... sweet.
    if (!forceStdIcon) {
        const icon = IEAppProjectIcon(containee)
        if (icon) {
            return icon
        }
    }
    return ContaineeTypeIcons[containee.flavor] || netnsIcon(containee)
}

const BusyboxIcons: { [key: string]: typeof SvgIcon | ((props: SvgIconProps) => JSX.Element) } = {
    ['ash']: TerminalIcon,
    ['bash']: TerminalIcon,
    ['csh']: TerminalIcon,
    ['dash']: TerminalIcon,
    ['fish']: TerminalIcon,
    ['ksh']: TerminalIcon,
    ['sh']: TerminalIcon,
    ['tsh']: TerminalIcon,
    ['zsh']: TerminalIcon,

    ['dumpcap']: CaptureIcon,
}

const netnsIcon = (containee: Containee) => {
    if (!isBusybox(containee)) {
        return ContaineeTypeIcons.netns
    }
    const pidsuffix = `(${containee.ealdorman.pid})`
    const name = containee.ealdorman.name.substring(0, containee.ealdorman.name.indexOf(pidsuffix))
    return BusyboxIcons[name] || ContaineeTypeIcons.netns
}

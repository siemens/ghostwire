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

import { isContainer, isPod, Containee, ContainerFlavors, PodFlavors } from 'models/gw'
import { IEAppProjectIcon } from './appicon'
import DockerManagedPluginIcon from 'icons/containees/DockerManagedPlugin'
import InitialIcon from 'icons/containees/Initial'


const ContaineeTypeIcons = {
    'netns': NetworkNamespaceIcon,
    'initialnetns': InitialIcon,
    'unknowntype': ContainerIcon,
    [ContainerFlavors.DOCKER]: DockerIcon,
    [ContainerFlavors.DOCKERPLUGIN]: DockerManagedPluginIcon,
    [ContainerFlavors.CONTAINERD]: ContainerdIcon,
    [ContainerFlavors.IERUNTIME]: IERuntimeIcon,
    [ContainerFlavors.IEAPP]: IEAppIcon,
    [ContainerFlavors.KIND]: KinDIcon,
}

const PodTypeIcons = {
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
        return (containee.netns && containee.netns.isInitial) ? ContaineeTypeIcons.initialnetns : ContaineeTypeIcons.netns
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
    return ContaineeTypeIcons[containee.flavor] || ContaineeTypeIcons.netns
}

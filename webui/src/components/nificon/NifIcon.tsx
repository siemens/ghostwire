// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { SvgIconProps } from '@mui/material'

import { NetworkInterface } from 'models/gw/nif'
import BridgeIcon from 'icons/nifs/Bridge'
import DummyIcon from 'icons/nifs/Dummy'
import MacvlanIcon from 'icons/nifs/Macvlan'
import TapIcon from 'icons/nifs/Tap'
import TunIcon from 'icons/nifs/Tun'
import VethIcon from 'icons/nifs/Veth'
import { BridgeInternalIcon, MacvlanMasterIcon, NicIcon, OverlayIcon } from 'icons/nifs'
import { GHOSTWIRE_LABEL_ROOT } from 'models/gw/model'
import { NifHWIcon } from 'components/nifhwicon'

// Known network interface type icons, indexed by the kind property of network
// interface objects (and directly taken from what Linux' RTNETLINK tells us).
const nifTypeIcons: { [key: string]: (props: SvgIconProps) => JSX.Element } = {
    'bridge': BridgeIcon,
    'dummy': DummyIcon,
    'macvlan': MacvlanIcon,
    'tap': TapIcon,
    'tun': TunIcon,
    'veth': VethIcon,
    'vxlan': OverlayIcon,
}

export interface NifIconProps extends SvgIconProps {
    /** network interface object describing a network interface in detail. */
    nif: NetworkInterface
    /** show HW NIC icon instead of generic icon if nic is "physical". */
    considerPhysical?: boolean
}

export const NifIcon = ({ nif, considerPhysical, ...props }: NifIconProps) => {
    if (!nif) {
        return <></>
    }
    if (considerPhysical && nif.isPhysical) {
        return <NifHWIcon nif={nif} {...props} />
    }
    if (nif.labels && GHOSTWIRE_LABEL_ROOT + 'bridge/internal' in nif.labels) {
        return <BridgeInternalIcon {...props} />
    }
    const Icon = (nif.tuntapDetails && nifTypeIcons[nif.tuntapDetails.mode]) ||
        (nif.macvlans && MacvlanMasterIcon) ||
        nifTypeIcons[nif.kind] || NicIcon
    return <Icon {...props} />
}

export default NifIcon

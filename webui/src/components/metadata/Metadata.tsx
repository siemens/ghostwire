// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { localStorageAtom } from 'utils/persistentsettings'
import { Collapse, IconButton, styled } from '@mui/material'
import { useDiscovery } from 'components/discovery'
import { useAtom } from 'jotai'
import { ExpandLess, ExpandMore } from '@mui/icons-material'

const expandMetadata = 'ghostwire.expandmd'

const expandMetadataAtom = localStorageAtom(expandMetadata, true)

const Section = styled('div')(({ theme }) => ({
    display: 'grid',
    marginLeft: theme.spacing(2),
}))

const MetadataTable = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'minmax(auto, max-content) minmax(50%, auto)', // https://stackoverflow.com/a/62163763
    columnGap: theme.spacing(2),
    rowGap: theme.spacing(0.5),

    margin: theme.spacing(1),
}))

const MetaName = styled('div')(({ theme }) => ({
    gridColumn: '1 / 2',
    color: theme.palette.text.secondary,
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    overflowWrap: 'break-word',
}))

const MetaValue = styled('div')(({ theme }) => ({
    gridColumn: '2 / 3',
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    overflowWrap: 'break-word',

    '& > .MuiSvgIcon-root': { verticalAlign: 'middle' },
}))

interface MetaRowProps {
    name: string
    value: any
}

const MetaRow = ({ name, value }: MetaRowProps) => {
    return (!!value &&
        <>
            <MetaName>{name}</MetaName>
            <MetaValue>{value}</MetaValue>
        </>)
        || null
}

/**
 * Component `Metadata` renders an expandable/collapsible division showing
 * discovery meta data. The expand/collapse state is made persistent in local
 * storage (and thus per "site").
 */
const Metadata = () => {
    const discovery = useDiscovery()
    const [expanded, setExpanded] = useAtom(expandMetadataAtom)

    if (!discovery.metadata) {
        return null
    }

    const hostos = [
        discovery.metadata['osrel-name'],
        discovery.metadata['osrel-version']
    ].join(" ")

    const iedmeta = discovery.metadata["industrial-edge"] || {}

    const coresemversion = iedmeta.semversion || undefined

    const engines = discovery.metadata["container-engines"]
        ? Object.values(discovery.metadata["container-engines"])
            .sort((engA, engB) => (engA["type"] + engA["version"]).localeCompare((engB["type"] + engB["version"])))
            .map((engine, idx) =>
                <MetaRow key={idx} name="Container engine" value={
                    `${engine["version"]} (type ${engine["type"]})`
                } />)
        : null

    const handleExpandClick = () => {
        setExpanded(!expanded)
    }

    return <div>
        <Section>
            <IconButton
                sx={{ justifySelf: 'center' }}
                edge="start"
                size="small"
                onClick={handleExpandClick}>
                {expanded ? <ExpandLess /> : <ExpandMore />}
            </IconButton>
            <Collapse in={expanded} mountOnEnter={true} timeout="auto">
                <MetadataTable>
                    <MetaRow name="IE device name" value={iedmeta['device-name']} />
                    <MetaRow name="Host name" value={discovery.metadata.hostname} />
                    <MetaRow name="Host OS" value={hostos} />
                    <MetaRow name="Kernel version" value={discovery.metadata['kernel-version']} />
                    <MetaRow name="Industrial Edge runtime" value={coresemversion} />
                    <MetaRow name="IE device developer mode" value={iedmeta['developer-mode'] === 'true' ? 'enabled' : undefined} />
                    {engines}
                </MetadataTable>
            </Collapse>
        </Section>
    </div>
}

export default Metadata

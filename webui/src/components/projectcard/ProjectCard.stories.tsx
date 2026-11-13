// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { NetnsPlainCard } from 'components/netnsplaincard'
import { type Container, GHOSTWIRE_LABEL_ROOT, isContainer, loFiNetns, type Project, ProjectFlavors } from 'models/gw'
import { crowdedNetns } from 'models/gw/mock'

import ProjectCard from './ProjectCard'

const netns = crowdedNetns!
const containers = netns.containers.filter(cntr => isContainer(cntr)) as Container[]

const composerProject: Project = {
    name: 'fooproject',
    flavor: ProjectFlavors.COMPOSER,
    containers: containers,
    netnses: { [netns.netnsid]: netns },
}

const ieappProject: Project = {
    name: 'myieapp',
    flavor: ProjectFlavors.IEAPP,
    containers: containers.map(cntr => ({
        ...cntr,
        flavor: ProjectFlavors.IEAPP,
        labels: { ...cntr.labels, [GHOSTWIRE_LABEL_ROOT + 'ieapp/title']: 'My IE App' },
    })),
    netnses: { [netns.netnsid]: netns },
}

const meta: Meta<typeof ProjectCard> = {
    title: 'Ghostwire/ProjectCard',
    component: ProjectCard,
    args: {
        project: composerProject,
        children: <NetnsPlainCard netns={loFiNetns(netns, true)} />,
    },
    argTypes: {
        project: { control: false },
        children: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ProjectCard>

export const ComposerProject: Story = {}

export const IEAppProject: Story = {
    args: {
        project: ieappProject,
    },
}

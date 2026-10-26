// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import ProjectCard from 'components/projectcard'
import { type Container, type Project, ProjectFlavors } from 'models/gw'
import { containerBox, podBox, standaloneBox } from 'views/help/mockdata'

import { Process } from './Process'

const project: Project = {
    name: 'fooproject',
    flavor: ProjectFlavors.COMPOSER,
    containers: [],
    netnses: {},
}

const projectContainer: Container = {
    ...containerBox,
    name: 'fooproject-web-1',
    labels: {
        'com.docker.compose.project': 'fooproject',
        'com.docker.compose.service': 'web',
    },
    project,
}

project.containers.push(projectContainer)

const meta: Meta<typeof Process> = {
    title: 'Ghostwire/Process',
    component: Process,
    args: {
        cmdline: ['/usr/bin/fool', '--world-domination'],
        pid: 12345,
        containee: containerBox,
    },
    argTypes: {
        containee: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Process>

export const InContainer: Story = {}

export const InPod: Story = {
    args: {
        containee: podBox,
    },
}

export const InComposerProjectContainer: Story = {
    args: {
        cmdline: ['/usr/sbin/nginx', '-g', 'daemon off;'],
        pid: 4242,
        containee: projectContainer,
    },
}

// The project isn't repeated inside its own project card.
export const InsideProjectCard: Story = {
    args: InComposerProjectContainer.args,
    decorators: [
        (Story) => <ProjectCard project={project}>
            <Story />
        </ProjectCard>,
    ],
}

export const StandaloneProcess: Story = {
    args: {
        cmdline: ['/sbin/init', 'splash'],
        pid: 1,
        containee: standaloneBox,
    },
}

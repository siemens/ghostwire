// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { MemoryRouter } from 'react-router'
import type { Meta, StoryObj } from '@storybook/react'
import { MuiMarkdown } from 'components/muimarkdown'

import { HelpViewer } from './HelpViewer'

import chintro from "./01-intro.mdx"
import chfoobar from "./02-foobar.mdx"
import chnew from "./03-newchapter.mdx"

const MyMarkdowner = (props: any) => (<MuiMarkdown {...props} />);

const chapters = [
    { title: "Intro", chapter: chintro },
    { title: "Foo Bar", chapter: chfoobar },
    { title: "A New Chapter", chapter: chnew },
];

const meta: Meta<typeof HelpViewer> = {
    title: 'Universal/HelpViewer',
    component: HelpViewer,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof HelpViewer>

export const Standard: Story = {
    render: () => <MemoryRouter initialEntries={['/help']}>
        <HelpViewer
            chapters={chapters}
            baseroute='/help'
            style={{ height: '30ex', maxHeight: '30ex' }}
            markdowner={MyMarkdowner}
        />
    </MemoryRouter>
}

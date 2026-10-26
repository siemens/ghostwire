// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box } from '@mui/material'

import { orderRoutes, RouteTableLocal, RouteType } from 'models/gw'
import { initialNetns } from 'models/gw/mock'

import { Route } from './Route'

// local table routes are never rendered, so leave them out from the start.
const routes = initialNetns.routes
    .filter(route => route.table !== RouteTableLocal)
    .sort(orderRoutes)

const defaultRoute = routes.find(route => route.prefixlen === 0 && !!route.nexthop) ?? routes[0]
const subnetRoute = routes.find(route => route.prefixlen !== 0 && !route.nexthop) ?? routes[0]

const meta: Meta<typeof Route> = {
    title: 'Ghostwire/Route',
    component: Route,
    args: {
        route: defaultRoute,
        onNavigation: fn(),
    },
    argTypes: {
        route: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Route>

export const DefaultRoute: Story = {}

export const SubnetRoute: Story = {
    args: {
        route: subnetRoute,
    },
}

export const Multicast: Story = {
    args: {
        route: { ...subnetRoute, destination: 'ff00::', prefixlen: 8, type: RouteType.Multicast, nexthop: undefined },
    },
}

export const AllRoutesOfHost: Story = {
    render: (args) => <>
        {routes.map((route, idx) => (
            <Box sx={{ py: 0.5 }} key={idx}>
                <Route {...args} route={route} />
            </Box>
        ))}
    </>,
}

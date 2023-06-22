// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

/// <reference types="../support" />

describe('ghostwire app', () => {

    before(() => {
        cy.log('loads')
        cy.visit('/')
        cy.waitForReact(2000, '#root')

        cy.log('refreshes')
        cy.react('Refresher')
            .find('button')
            .first().click()
        cy.getReact('NamespaceInfo')
            .nthNode(0)
            .getProps('namespace').then((netns) => {
                expect(netns).has.property('type', 'user')
                expect(netns).has.property('initial', true)

            })
    })

    it('shows about', () => {
        cy.historyPush('/about')
        cy.react('About').contains('Version')
    })

    it('lends a helping hand', () => {
        cy.historyPush('/help')
        cy.react('HelpViewer').contains('The information in this help')
    })

})

export { }

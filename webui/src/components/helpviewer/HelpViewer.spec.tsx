// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

/// <reference types="../../cypress/support" />

import React from 'react'
import { MemoryRouter } from 'react-router-dom'
import { CypressHistorySupport } from 'cypress-react-router'
import { mount } from '@cypress/react'

import chintro from "!babel-loader!mdx-loader!./01-intro.mdx"
import chfoobar from "!babel-loader!mdx-loader!./02-foobar.mdx"
import chnew from "!babel-loader!mdx-loader!./03-newchapter.mdx"
import HelpViewer from './HelpViewer'

const chapters = [
    { title: "Intro", chapter: chintro },
    { title: "Foo Bar", chapter: chfoobar },
    { title: "A New Chapter", chapter: chnew },
]

describe('HelpViewer', () => {

    it('helps', () => {
        mount(
            <MemoryRouter initialEntries={['/foo/help']}>
                <CypressHistorySupport />
                <HelpViewer chapters={chapters} baseroute='/foo/help' />
            </MemoryRouter>
        )
        cy.waitForReact()
        cy.react('HelpViewer').should('exist').as('hv')
        cy.react('IconButton').should('exist').as('nav')

        cy
            .get('@hv').find('h4').contains('Introduction')
            .get('button.prev').should('not.exist')
            .get('button.next').contains('Foo Bar')
            .click()

        cy
            .history().its('location').its('pathname').should('equal', '/foo/help/foobar')
            .get('@hv').find('h4').contains('Foo Bar')
            .get('button.prev').contains('Intro')
            .get('button.next').contains('A New Chapter')
            .click()

        cy
            .get('@hv').find('h4').contains('A New chapter')
            .get('button.prev').contains('Foo Bar')
            .get('button.next').should('not.exist')

        cy
            .get('@nav').click()
            .get('.MuiPopover-paper').find('li').each((navitem, idx) => {
                expect(navitem.text()).to.equal(chapters[idx].title)
                if (idx === chapters.length - 1) {
                    expect(navitem).to.have.class('Mui-selected')
                }
            })
            .first().click()

        cy
            .get('@hv').find('h4').contains('Introduction')
            .history().its('location').its('pathname').should('equal', '/foo/help/intro')
    })


})

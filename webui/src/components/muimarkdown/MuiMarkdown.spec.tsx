// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { ComponentType } from 'react'
import { mount } from '@cypress/react'
import { MuiMarkdown } from './MuiMarkdown'
import pDefer from 'p-defer'

import TestMDX from "!babel-loader!mdx-loader!./MuiMarkdown.spec.mdx"


describe('MuiMarkdown', () => {

    it('renders synchronous MDX', () => {
        mount(<MuiMarkdown mdx={TestMDX} />)
        cy.waitForReact()
        cy.get('#headah')
            .should('have.length', 1)
            .contains('Headah')
        cy.get('strong').contains('text')
    })

    it('renders lazy MDX with default fallback', () => {
        const deferredImportPromise = pDefer()
        const deferredMDX = React.lazy(() =>
            (deferredImportPromise.promise as Promise<{ default: ComponentType<any> }>))

        mount(<MuiMarkdown mdx={deferredMDX} />)
        cy.waitForReact()
        cy.get('.MuiSkeleton-root').should('exist')

        cy.then(() => deferredImportPromise.resolve({ default: TestMDX }))
            .get('#headah')
            .should('have.length', 1)
            .contains('Headah')
            // fallback skeleton should be gone by now.
            .get('.MuiSkeleton-root', { timeout: 100 }).should('not.exist')
    })

    it('renders custom fallback', () => {
        const deferredImportPromise = pDefer()
        const deferredMDX = React.lazy(() =>
            (deferredImportPromise.promise as Promise<{ default: ComponentType<any> }>))

        const MyFallback = () => <span id="myfallback">myfallback</span>

        mount(<MuiMarkdown mdx={deferredMDX} fallback={<MyFallback />} />)
        cy.waitForReact()
        cy.get('#myfallback').should('exist')
            .contains('myfallback')

        cy.then(() => {deferredImportPromise.reject()})
    })

})
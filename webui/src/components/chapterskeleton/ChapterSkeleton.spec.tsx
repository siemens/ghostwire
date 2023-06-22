// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { mount } from '@cypress/react'
import { ChapterSkeleton } from './ChapterSkeleton'

describe('ChapterSkeleton', () => {

    it('renders', () => {
        mount(
            <ChapterSkeleton sx={{width: '10rem'}} />
        )
        cy.waitForReact()
        cy.get('.MuiTypography-h4')
            .should('have.length', 1)
            .find('.MuiSkeleton-root')
        cy.get('.MuiTypography-body1')
            .should('have.length', 3)
            .find('.MuiSkeleton-root')
    })

})
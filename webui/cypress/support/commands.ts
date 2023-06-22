// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

// See also:
// https://github.com/omerose/cypress-support/blob/master/cypress/support/commands.ts

Cypress.Commands.add(
    'history',
    () => cy.window().its('cyHistory')
)

// Make this a module.
export { }

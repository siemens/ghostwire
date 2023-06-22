// Derived from https://docs.cypress.io/guides/tooling/plugins-guide
//
// Copyright (c) 2022 Cypress.io
//
// SPDX-License-Identifier: MIT
//
// Modifications by Siemens

module.exports = (on, config) => {
  require('@cypress/react/plugins/react-scripts')(on, config)
  return config
}

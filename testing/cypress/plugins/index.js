/// <reference types="cypress" />
// ***********************************************************
// This example plugins/index.js can be used to load plugins
//
// You can change the location of this file or turn off loading
// the plugins file with the 'pluginsFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/plugins-guide
// ***********************************************************

// This function is called when a project is opened or re-opened (e.g. due to
// the project's config changing)

/**
 * @type {Cypress.PluginConfig}
 */
// eslint-disable-next-line no-unused-vars
module.exports = (on, config) => {
  // modify the config values
  config.defaultCommandTimeout = 20000
 
  config.env.repoUrl = "git@github.com:caos/cypress-ops.git"
  config.env.orbConfig = "./orbconfig"
  config.env.orbCtl = "./orbctl"
  config.env.testPassword = "somesortofcrypticpw"
  config.env.releaseVersion ="https://api.github.com/repos/caos/orbos/releases/tags/cypress-testing-dev"
//  config.env.releaseVersion ="https://api.github.com/repos/caos/orbos/releases/latest"
  // IMPORTANT return the updated config object
  return config

}

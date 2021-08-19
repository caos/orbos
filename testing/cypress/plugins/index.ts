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

//import "cypress-fail-fast";

// This function is called when a project is opened or re-opened (e.g. due to
// the project's config changing)

export type LogEntry = {
  msg: string
  origin: string
}

/**
 * @type {Cypress.PluginConfig}
 */
// eslint-disable-next-line no-unused-vars
module.exports = (on, config) => {

  // modify the config values
  config.defaultCommandTimeout = 20000
  config.screenshotOnRunFailure = false
  config.video = false

  config.env.e2eOrbUrl = process.env.E2E_ORB_URL
  config.env.orbosTag = process.env.ORBOS_TAG
  config.env.githubAccessToken = process.env.GITHUB_ACCESS_TOKEN
  config.env.workFolder = './work'
  config.env.orbctlFile = `${config.env.workFolder}/orbctl`
  config.env.orbconfigFile = `${config.env.workFolder}/orbconfig`
  config.env.orbctl = `${config.env.orbctlFile} --disable-analytics`
  config.env.orbctlGitops = `${config.env.orbctl} --gitops --orbconfig ${config.env.orbconfigFile}`
  config.env.retrySeconds = 10

  on('task', {
    info(entry: LogEntry) {
      console.info(`${entry.origin} info:`, entry.msg)
      return null
    },
    error(entry: LogEntry) {
      console.info(`${entry.origin} error:`, entry.msg)
      return null
    },
    env(key: string): string {
      return process.env[key] || ""
    },
  })
  return config  

}

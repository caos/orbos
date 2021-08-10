
import cypress = require("cypress")
import { load as ymlToObj, dump as objToYML} from "js-yaml"

declare global {
    namespace Cypress {
        interface Chainable<Subject> {
            /**
             * Custom command that unlike cy.exec throws an error including the commands full stdout and stderr.
             * see https://github.com/cypress-io/cypress/issues/5470#issuecomment-569627930
             * 
             * @example cy.execZero('whoami', result => { expect(result.stdout).to.be('hodor') })
             */
            execZero(command: string): Chainable<Exec>

            /**
             * Custom orbctl command with analytics turned off
             * 
             * @example cy.orbctl('version')
             */
            orbctl(args: string): Chainable<Exec>

            /**
             * Custom orbctl command with standard GitOps flags appended and analytics turned off
             * 
             * @example cy.orbctlGitops('version')
             */
            orbctlGitops(args: string): Chainable<Exec>

            /**
             * Custom command that unmarshals a YAML string into a JSON object
             * 
             * @example cy.toObject('version')
             */
             toObject(): Chainable<any>

            /**
             * Custom command that marshals a JSON object into a YAML string
             * 
             * @example cy.toObject('version')
             */
             toYAML(): Chainable<string>
        }
    }
}

Cypress.Commands.add('execZero', { prevSubject: false }, (command: string) =>
    cy.exec(command, { timeout: 60 * 1000, failOnNonZeroExit: false }).then(result => {
        if (result.code) {
            throw new Error(`Execution of "${command}" failed
Exit code: ${result.code}
Stdout:\n${result.stdout}
Stderr:\n${result.stderr}`);
        }
        return result
    })
)

Cypress.Commands.add('orbctl', { prevSubject: false }, (args: string) =>
    cy.execZero(`${Cypress.env('orbctl')} --disable-analytics ${args}`))

Cypress.Commands.add('orbctlGitops', { prevSubject: false }, (args: string) =>
    cy.orbctl(`--gitops --orbconfig ${Cypress.env('orbconfig')} ${args}`))

Cypress.Commands.add('toObject', { prevSubject: true }, (yml: string) => cy.wrap(ymlToObj(yml)))

Cypress.Commands.add('toYAML', { prevSubject: true }, (obj: any) => cy.wrap(objToYML(obj)))
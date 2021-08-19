
import cypress = require("cypress")
import { load as ymlToObj, dump as objToYML} from "js-yaml"
import { LogEntry } from "../plugins";

declare global {
    namespace Cypress {
        interface Chainable<Subject> {
            /**
             * Overwritten command that unlike the original cy.exec throws an error including the commands full stdout and stderr.
             * see https://github.com/cypress-io/cypress/issues/5470#issuecomment-569627930
             * 
             * @example cy.execZero('whoami', result => { expect(result.stdout).to.be('hodor') })
             */
            exec(command: string, options?: Partial<Cypress.ExecOptions>): Chainable<Exec>

            /**
             * Custom command that unmarshals a YAML string into a JSON object
             * 
             * @example cy.toObject('some: yaml')
             */
             toObject(): Chainable<any>

            /**
             * Custom command that marshals a JSON object into a YAML string
             * 
             * @example cy.toYAML({"some": 'json'})
             */
             toYAML(json?: any): Chainable<string>

             /**
             * Custom command that logs the passed error and stops the cypress runner
             * 
             * @example cy.panic('a fatal error occurred')
             */
             panic(err: string): void
        }
    }
}

function panic(err: string): void {
    cy.task('error', <LogEntry>{msg: err, origin: 'cypress'}).then(() => {
        expect(false).to.equal(true)
        Cypress.runner.stop()
    })
}

Cypress.Commands.add('panic', { prevSubject: false }, (err: string) => {
    panic(err)
})

Cypress.Commands.overwrite('exec', (originalFn: (...args: any[]) => Cypress.Chainable<Cypress.Exec>, command: string, options?: Partial<Cypress.ExecOptions>) => {

    const mustSucceed = !options || options.failOnNonZeroExit === undefined || options.failOnNonZeroExit

    return originalFn(command, Object.assign(options || {}, { failOnNonZeroExit: false })).then(result => {
        if (result.code && mustSucceed) {
            cy.task('error', <LogEntry>{msg: `Execution of "${command}" failed
Exit code: ${result.code}
Stdout:\n${result.stdout}
Stderr:\n${result.stderr}`, origin: 'cypress'}).then(() => {
                expect(false).to.equal(true)
                Cypress.runner.stop()
            })
        
/*            panic(`Execution of "${command}" failed
Exit code: ${result.code}
Stdout:\n${result.stdout}
Stderr:\n${result.stderr}`)*/
        }
        return result
    })
})

Cypress.Commands.add('toObject', { prevSubject: true }, (yml: string) => cy.wrap(ymlToObj(yml)))

Cypress.Commands.add('toYAML', { prevSubject: 'optional' }, (subj: any, arg: any) => cy.wrap(objToYML(subj ? subj : arg)))

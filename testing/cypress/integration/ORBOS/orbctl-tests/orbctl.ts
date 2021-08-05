import { match } from "cypress/types/sinon";
import { load as ymlToObj} from "js-yaml"

const orbctlGitops = `${Cypress.env("orbCtl")} --gitops --orbconfig ${Cypress.env("orbConfig")} --disable-analytics`
const retrySeconds = 10

interface errorMeta {
    operator: string     
    since: string     
}

// Stops the runner on the first failing test and waits between retries
afterEach(function() {

    if (!this.currentTest.isFailed()) {
        return
    }

    const totalRetries = this.currentTest.retries()
    if (!totalRetries) {
        handleError(this.currentTest.err, null)
        return
    }

    const remaining = this.remainingRetries === undefined ? totalRetries : this.remainingRetries-1
    cy.wrap(remaining).as("remainingRetries").then(remainingRetries => {
        cy.task('info', `retrying ${remainingRetries}, totally ${totalRetries} times`)
        if (remainingRetries < 1) {
            cy.wrap(undefined).as("remainingRetries")
            const errorMeta = <errorMeta>this.errorMeta
            handleError(this.currentTest.err, () => {
                cy.exec(`kubectl -n caos-system logs "app.kubernetes.io/name=${errorMeta.operator}" --since-time ${errorMeta.since}`).then(result => {
                    
                })
            })
        }
        cy.wait(retrySeconds * 1000)
    })
});

describe('install orbctl', { execTimeout: 90000 }, () => {
    // prepare orbctl, download and configuration
    it('download orbctl', () => {
        cy.exec('bash -c \"if [ -f "./orbctl" ]; then rm ./orbctl ; fi\"').its('code').should('eq', 0);
        cy.exec(`curl -s ${Cypress.env("releaseVersion")} | grep "browser_download_url.*orbctl.$(uname).$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ./orbctl `).its('code').should('eq', 0);
        cy.exec('[ -s ./orbctl ]').its('code').should('eq', 0);
        //TODO: check filesize > 0
    });
    it('chmod orbctl', () => {
        cy.exec('chmod +x ./orbctl ').its('code').should('eq', 0);
    })
});

describe('initalize repo', () => {
/*    it('should initialize the orbiter.yml', () => {
        cy.exec(`${orbctlGitops} file patch orbiter.yml --exact --value ${to}`, { timeout: 20000 }).then(result => {
            cy.log(result.stdout)
            cy.log(result.stderr) 
        }).its('code').should('eq', 0)
    });
*/
    it('orbctl configure', () => {
        cy.exec(`${orbctlGitops} configure --repourl ${Cypress.env("repoUrl")} --masterkey "$(openssl rand -base64 21)"`, { timeout: 60 * 1000 }).then(result => {
            cy.log(result.stdout);
            cy.log(result.stderr);
        }).its('code').should('eq', 0);
    });
});
//TODO: GH token shall be set and automatically renewed


describe('orbctl tests', { execTimeout: 90000 }, () => {
    // basic orbctl operations
    it('orbctl help', () => {
        cy.exec(`${orbctlGitops} help`, { timeout: 20000 }).then(result => {
            //return JSON.parse(result.stdout)
            cy.log(result.stdout);
            cy.log(result.stderr);
            //return String.parse(result.stdout)
        }).its('code').should('eq', 0);
    });

    it('orbctl writesecret', () => {
        cy.exec(`${orbctlGitops} writesecret boom.monitoring.admin.password.encrypted --value ${Cypress.env("testPassword")}`, { timeout: 20000 }).then(result => {
            cy.log(result.stdout);
            cy.log(result.stderr);
        }).its('code').should('eq', 0);
    });

    it('orbctl readsecret', () => {
        cy.exec(`${orbctlGitops} readsecret boom.monitoring.admin.password.encrypted`, { timeout: 20000 }).then(result => {
            return result.stdout;
        }).then(returnpw => { expect(returnpw).to.eq(Cypress.env("testPassword")) });
    })
});

/*
describe('orbctl cleanup yml', { execTimeout: 90000 }, () => {

    it('orbctl remove/patch secret', () => {
        cy.exec(`${orbctlGitops} file patch --file boom.yml spec.monitoring.admin.password.value --exact --value ""`, { timeout: 20000 }).then(result => {
            cy.log(result.stdout)
            cy.log(result.stderr)
        }).its('code').should('eq', 0)
    })
})
*/

describe('orbctl takeoff', { execTimeout: 900000 }, () => {
    // orbctl takeoff
    it('orbctl --gitops takeoff', () => {
        cy.exec(`${orbctlGitops} takeoff`).its('code').should('eq', 0)
    });
});

function scale(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, pool: string, to: number, expectTotal: number, ensureTimeoutSeconds: number) {

    describe(`init scaling ${pool}`, {execTimeout: 90000}, () => {
        it(`orbctl patch should set the desired scale ${pool} to ${to}`, () => {
            cy.exec(`${orbctlGitops} file patch orbiter.yml clusters.orbos-test-gceprovider.spec.${pool}.nodes --exact --value ${to}`, { timeout: 20000 })
                .its('code')
                .should('eq', 0)
        })
    })

    describe(`scaling ${pool}`, { execTimeout: 90000, retries: Math.ceil(ensureTimeoutSeconds / retrySeconds) }, () => {
        it(`cluster should have ${expectTotal} machines`, function() {
            const since = new Date()       
            cy.wrap(<errorMeta>{
                operator: "",
                since: since.toISOString() 
            }).as('errorMeta').exec(`${orbctlGitops} file print caos-internal/orbiter/current.yml`, { timeout: 20000 })
                .then(result => {
                expect(result.code).to.equal(0)
                interface CurrentOrbiter { clusters: { "orbos-test-gceprovider" : { current: { status: string, machines: object} } } }
                const currentOrbiter: CurrentOrbiter = <any>ymlToObj(result.stdout);
                const currentCluster = currentOrbiter.clusters["orbos-test-gceprovider"].current
                expect(Object.keys(currentCluster.machines).length).to.eq(expectTotal)
                expect(currentCluster.status).to.eq("running")
            })
            // TODO check nodes with kubectl
        })
    })
}

scale(describe, "workers.0", 3, 4, 15 * 60)
scale(describe, "workers.0", 1, 2, 5 * 60)
scale(describe.only, "controlplane", 2, 3, 15 * 60)
scale(describe, "controlplane", 1, 2, 5 * 60)

function handleError(err: any, onError: () => string) {
    cy.task('error', err)
    if (onError) {
        const anotherErr = onError()
        if (anotherErr) {
            cy.task('error', anotherErr)
        }
    }
    Cypress.runner.stop()
}










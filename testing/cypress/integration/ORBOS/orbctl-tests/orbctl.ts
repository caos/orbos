import { load as ymlToObj} from "js-yaml"
import { LogEntry } from "../../../plugins";

const orbctlGitops = `${Cypress.env("orbCtl")} --gitops --orbconfig ${Cypress.env("orbConfig")} --disable-analytics`
const retrySeconds = 2

interface errorMeta {
    operator: string     
    since: string     
}

// Stops the runner on the first failing test and waits between retries
afterEach(function() {

    const ctx = this

    if (!ctx.currentTest.isFailed()) {
        cy.wrap(undefined).as("retryMeta")
        return
    }

    const panic = (err: any) => {
        cy.task('error', <LogEntry>{msg: err, origin: 'cypress'})
        Cypress.runner.stop()
    }

    const totalRetries = ctx.currentTest.retries()
    if (!totalRetries) {
        panic(ctx.currentTest.err)
    }

    interface RetryMeta {
        remaining: number
        since: string
    }
    const currentRetryMeta: RetryMeta = ctx.retryMeta
    const newRetryMeta: RetryMeta = {
        remaining: currentRetryMeta === undefined ? totalRetries : currentRetryMeta.remaining-1,
        since: currentRetryMeta === undefined ? new Date().toISOString() : currentRetryMeta.since
    }
    cy.wrap(newRetryMeta).as("retryMeta").then(meta => {
        if (meta.remaining > 0) {
            cy.task('info', <LogEntry>{msg: `retrying ${meta.remaining}, totally ${totalRetries} times`, origin: 'cypress'})
            cy.wait(retrySeconds * 1000)
            return
        }
        cy.exec(`kubectl -n caos-system logs --selector "app.kubernetes.io/name=${ctx.currentTest.ctx.operator}" --since-time ${meta.since}`).then(result => {
            return result.stdout.split("\n").forEach((entry) => {
                var level = 'info'
                if (entry.indexOf(" err=") > -1) {
                    level = 'error'
                }
                cy.task(level, <LogEntry>{msg: entry, origin: ctx.currentTest.ctx.operator})
            })
        })
        panic(ctx.currentTest.err)
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
            cy.wrap("orbiter").as('operator').exec(`${orbctlGitops} file print caos-internal/orbiter/current.yml`, { timeout: 20000 }).then(result => {
                expect(result.code, 'print current.yml return code').to.equal(0)
                interface CurrentOrbiter { clusters: { "orbos-test-gceprovider" : { current: { status: string, machines: object} } } }
                const currentOrbiter: CurrentOrbiter = <any>ymlToObj(result.stdout);
                const currentCluster = currentOrbiter.clusters["orbos-test-gceprovider"].current
                expect(Object.keys(currentCluster.machines).length, `${pool} machines`).to.eq(expectTotal)
                expect(currentCluster.status, 'cluster status').to.eq("running")
            })
            // TODO check nodes with kubectl
        })
    })
}

scale(describe, "workers.0", 3, 4, 15 * 60)
scale(describe, "workers.0", 1, 2, 5 * 60)
scale(describe.only, "controlplane", 2, 3, /*15 * 60*/ 2 * 60)
scale(describe, "controlplane", 1, 2, 5 * 60)










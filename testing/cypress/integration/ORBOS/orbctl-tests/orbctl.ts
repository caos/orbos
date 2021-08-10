import { LogEntry } from "../../../plugins";

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
            cy.wait(Cypress.env('retrySeconds') * 1000)
            return
        }
        cy.execZero(`kubectl -n caos-system logs --selector "app.kubernetes.io/name=${ctx.currentTest.ctx.operator}" --since-time ${meta.since}`).then(result => {
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
    it('should print the correct version', () => {
        cy.execZero(`bash -c \"if [ -f "${Cypress.env('orbctl')}" ]; then rm ${Cypress.env('orbctl')} ; fi\"`);
        cy.execZero(`curl -s https://api.github.com/repos/caos/orbos/releases/tags/${Cypress.env('orbosTag')} | grep "browser_download_url.*orbctl.$(uname).$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ${Cypress.env('orbctl')}`);
        cy.execZero(`chmod +x ${Cypress.env('orbctl')}`);
        cy.orbctl(`--version`).then(result => {
            expect(result.stdout.split(" ")[2]).to.equal(Cypress.env('orbosTag'))
        })
    });
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
        cy.writeFile(Cypress.env('orbconfig'), '')
        cy.writeFile('ghtoken', `IDToken: ""
IDTokenClaims: null
access_token: ${Cypress.env('githubAccessToken')}
expiry: "0001-01-01T00:00:00Z"
token_type: bearer`)
        cy.orbctlGitops(`configure --repourl ${Cypress.env("e2eOrbUrl")} --masterkey "$(openssl rand -base64 21)"`)
    });
});

describe('orbctl tests', { execTimeout: 90000 }, () => {
   
    it('orbctl help', () => {
        cy.orbctlGitops(`help`);
    });



    it('orbctl writesecret', () => {
        cy.orbctlGitops(`writesecret boom.monitoring.admin.password.encrypted --value dummy`);
    });

    it('orbctl readsecret', () => {
        cy.orbctlGitops(`readsecret boom.monitoring.admin.password.encrypted`)
    })
});

describe('orbctl takeoff', { execTimeout: 900000 }, () => {
    // orbctl takeoff
    it('orbctl --gitops takeoff', () => {
        cy.orbctlGitops(`takeoff`)
    });
});

scale(describe, "workers.0", 3, 4, 15 * 60)
scale(describe, "workers.0", 1, 2, 5 * 60)
scale(describe, "controlplane", 3, 4, 15 * 60)
scale(describe, "controlplane", 1, 2, 5 * 60)


function scale(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, pool: string, to: number, expectTotal: number, ensureTimeoutSeconds: number) {

    describe(`init scaling ${pool}`, {execTimeout: 90000}, () => {
        it(`orbctl patch should set the desired scale ${pool} to ${to}`, () => {
            cy.orbctlGitops(`file patch orbiter.yml clusters.orbos-test-gceprovider.spec.${pool}.nodes --exact --value ${to}`)
        })
    })

    describe(`scaling ${pool}`, { execTimeout: 90000, retries: Math.ceil(ensureTimeoutSeconds / Cypress.env('retrySeconds')) }, () => {
        it(`cluster should have ${expectTotal} machines`, function() {
            cy.orbctlGitops(`file print caos-internal/orbiter/current.yml`).then(result => {
                interface CurrentOrbiter { clusters: { "orbos-test-gceprovider" : { current: { status: string, machines: object} } } }
                const currentOrbiter: CurrentOrbiter = <any>ymlToObj(result.stdout);
                const currentCluster = currentOrbiter.clusters["orbos-test-gceprovider"].current
                expect(Object.keys(currentCluster.machines).length, `${pool} machines`).to.eq(expectTotal)
                expect(currentCluster.status, 'cluster status').to.eq("running")
                // TODO check nodes with kubectl
            })
        })
    })
}

describe.only('debug', () => {
    it('debug', () => {
        cy.fixture('orbiter-init.yml').toObject().then(orbiter => {
            cy.fixture('clusterspec-init.yml').toObject().then(cluster => {
                cluster.versions.orbiter = Cypress.env('orbosTag')
                orbiter.clusters.k8s.spec = cluster
                cy.orbctlGitops('file print provider.yml').toObject().then(provider => {
                    return Object.assign(cluster.providers.providerundertest, provider)
                }).toYAML().then(composed => {
                    cy.task('info', <LogEntry>{msg: composed, origin: 'debug'})
                })
            })
        })
    })
})
/*
.toYAML().then(composed => {
    cy.task('info', <LogEntry>{msg: composed, origin: 'debug'})
})
*/
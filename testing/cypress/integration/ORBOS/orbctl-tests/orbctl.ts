import { LogEntry } from "../../../plugins";

// Stops the runner on the first failing test and waits between retries
afterEach(stopAndReport);

describe('install orbctl', () => {
    // prepare orbctl, download and configuration
    it('should print the correct version', () => {
        cy.exec(`bash -c \"if [ -f "${Cypress.env('orbctlFile')}" ]; then rm ${Cypress.env('orbctlFile')} ; fi\"`);
        cy.exec(`curl -s https://api.github.com/repos/caos/orbos/releases/tags/${Cypress.env('orbosTag')} | grep "browser_download_url.*orbctl.$(uname).$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ${Cypress.env('orbctlFile')}`);
        cy.exec(`chmod +x ${Cypress.env('orbctlFile')}`);
        cy.exec(`${Cypress.env('orbctl')} --version`).then(result => {
            expect(result.stdout.split(" ")[2]).to.equal(Cypress.env('orbosTag'))
        })
    });
});

describe('initalize repo', () => {

    afterEach(stopAndReport);

    it('configures repo access', () => {
        cy.writeFile(Cypress.env('orbconfigFile'), '')
        cy.writeFile(`${Cypress.env('workFolder')}/ghtoken`, `IDToken: ""
IDTokenClaims: null
access_token: ${Cypress.env('githubAccessToken')}
expiry: "0001-01-01T00:00:00Z"
token_type: bearer`)
        cy.exec(`${Cypress.env('orbctlGitops')} configure --repourl ${Cypress.env("e2eOrbUrl")} --masterkey "$(openssl rand -base64 21)"`, { failOnNonZeroExit: false })
        cy.exec(`${Cypress.env('orbctlGitops')} file remove orbiter.yml boom.yml`)
    });

    describe('configure repo content', () => {

        before('cleanup files and load config', () => {
            cy.exec(`${Cypress.env('orbctlGitops')} file print e2e.yml`).its("stdout").toObject().as('e2e')
            cy.fixture('boom-init.yml').toObject().then(desiredBoom => {
                desiredBoom.spec.boomVersion = Cypress.env('orbosTag')
                return desiredBoom
            }).as('desiredBoom')
        })

        afterEach(stopAndReport);

        describe('with loaded configuration', () => {

            it('creates boom.yml configuration file', function() {
                cy.toYAML(this.desiredBoom).then(boomYml => {
                    cy.exec(`${Cypress.env('orbctlGitops')} file patch boom.yml --exact --value '${boomYml}'`)
                })
            })

            it('creates orbiter.yml configuration file', () => {
                cy.fixture('orbiter-init.yml').toObject().then(orbiter => {
                    cy.fixture('clusterspec-init.yml').toObject().then(cluster => {
                        cluster.versions.orbiter = Cypress.env('orbosTag')
                        orbiter.clusters.k8s.spec = cluster
                        cy.exec(`${Cypress.env('orbctlGitops')} file print provider-init.yml`).its("stdout").toObject().then(provider => {
                            orbiter.providers.providerundertest = provider
                            return orbiter
                        }).then(orbiter => {
                            cy.fixture('loadbalancing-init.yml').toObject().then(loadbalancing => {
                                orbiter.providers.providerundertest.loadbalancing = loadbalancing
                                return orbiter
                            }).toYAML().then(composed => {
                                cy.writeFile(`${Cypress.env('workFolder')}/orbiter.yml`, composed).then(() => {
                                    cy.exec(`${Cypress.env('orbctlGitops')} file patch orbiter.yml --exact --file ${Cypress.env('workFolder')}/orbiter.yml`)
                                })
                            })
                        })
                    })
                })
            })

            it('migrates the api', () => {
                cy.exec(`${Cypress.env('orbctlGitops')} api`)
            })

            it('writes provider secrets', function() {
                Object.entries(this.e2e.initsecrets).forEach((secretRef) => {
                    cy.task('env', secretRef[1]).then(value => {
                        cy.exec(`${Cypress.env('orbctlGitops')} writesecret orbiter.providerundertest.${secretRef[0]}.encrypted --value '${value}'`)
                    })
                })
            });

            it('generates some secrets', () => {
                cy.exec(`${Cypress.env('orbctlGitops')} configure`)
            })
        })
    })
});

describe('orbctl tests', () => {
   
    it('orbctl help', () => {
        cy.exec(`${Cypress.env('orbctl')} help`);
    });

    it('orbctl writesecret', () => {
        cy.exec(`${Cypress.env('orbctlGitops')} writesecret boom.monitoring.admin.password.encrypted --value dummy`);
    });

    it('orbctl readsecret', () => {
        cy.exec(`${Cypress.env('orbctlGitops')} readsecret boom.monitoring.admin.password.encrypted`).its('stdout').should('equal', 'dummy')
    })
});

execLongRunning(describe, 'should successfully return when the cluster is bootstrapped and ORBITER took off', `${Cypress.env('orbctlGitops')} takeoff`, 15 * 60 * 1000)

describe('bootstrap orb', () => {
    // orbctl takeoff
    it('should successfully return when the cluster is bootstrapped and ORBITER took off', () => {
        cy.exec(`${Cypress.env('orbctlGitops')} takeoff`, { timeout: 15 * 60 * 1000 })
    });
});

scale(describe, "workers.0", 3, 4, 15 * 60)
scale(describe, "workers.0", 1, 2, 5 * 60)
scale(describe, "controlplane", 3, 4, 15 * 60)
scale(describe, "controlplane", 1, 2, 5 * 60)


function scale(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, pool: string, to: number, expectTotal: number, ensureTimeoutSeconds: number) {

    describe(`init scaling ${pool}`, () => {
        it(`orbctl patch should set the desired scale ${pool} to ${to}`, () => {
            cy.exec(`${Cypress.env('orbctlGitops')} file patch orbiter.yml clusters.k8s.spec.${pool}.nodes --exact --value ${to}`)
        })
    })

    describe(`scaling ${pool}`, { retries: Math.ceil(ensureTimeoutSeconds / Cypress.env('retrySeconds')) }, () => {
        it(`cluster should have ${expectTotal} machines`, function() {
            cy.exec(`${Cypress.env('orbctlGitops')} file print caos-internal/orbiter/current.yml`).its('stdout').toObject().its('clusters.k8s.current').then(current => {
                expect(Object.keys(current.machines).length, `${pool} machines`).to.equal(expectTotal)
                expect(current.status, 'cluster status').to.equal("running")
                // TODO check nodes with kubectl
            })
        })
    })
}

execLongRunning(describe, 'destroys the orb', `yes | ${Cypress.env('orbctlGitops')} destroy`, 5 * 60 * 1000)

function stopAndReport() {

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
}

function execLongRunning(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, itMsg: string, command: string, timeoutMS: number, then?: (chainable: Cypress.Chainable<Cypress.Exec>) => void) {

    describe(`setup command: ${command}`, () => {
        after(stopAndReport);

        after(function(){
           if (this.currentTest.err.message.indexOf('timed out after waiting ') > -1) {
                cy.readFile(`${Cypress.env('workFolder')}/longrunning.log`).then(log => 
                    cy.task('error', <LogEntry>{msg: `Command timed out.
${log}`, origin: 'cypress'}))
            }
        })
    
        describe(`execute command: ${command}`, () => {
            it(itMsg, () => {
                const c = cy.exec(`bash -c '${command}' > ${Cypress.env('workFolder')}/longrunning.log`, { timeout: timeoutMS })
                if (then) {
                    then(c)
                }
            })
        })
    })
}


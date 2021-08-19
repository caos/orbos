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

    before('initialize local files', () => {
        cy.writeFile(Cypress.env('orbconfigFile'), '')
        cy.writeFile(`${Cypress.env('workFolder')}/ghtoken`, `IDToken: ""
IDTokenClaims: null
access_token: ${Cypress.env('githubAccessToken')}
expiry: "0001-01-01T00:00:00Z"
token_type: bearer`)
    })

    execLongRunning(describe, 'should create the orbconfig and add a new deploy key to the repo', `${Cypress.env('orbctlGitops')} configure --repourl ${Cypress.env("e2eOrbUrl")} --masterkey "$(openssl rand -base64 21)"`, 'orbctl', { timeout: 5 * 60 * 1000, failOnNonZeroExit: false })

    describe('initialize repo', () => {

        afterEach(stopAndReport);

        before('load configuration', () => {
            cy.exec(`${Cypress.env('orbctlGitops')} file print e2e.yml`).its("stdout").toObject().as('e2e')
            cy.fixture('boom-init.yml').toObject().then(desiredBoom => {
                desiredBoom.spec.boomVersion = Cypress.env('orbosTag')
                return desiredBoom
            }).as('desiredBoom')
        })

        describe('with loaded configuration', () => {

            it('should have loaded the configuration files', function() {
                expect(this.desiredBoom).to.not.be.undefined
            })

            it('should remove remote files', () => {
                cy.exec(`${Cypress.env('orbctlGitops')} file remove orbiter.yml boom.yml`)
            })

            it('should create the raw boom.yml configuration file', function() {
                cy.toYAML(this.desiredBoom).then(boomYml => {
                    cy.exec(`${Cypress.env('orbctlGitops')} file patch boom.yml --exact --value '${boomYml}'`)
                })
            })

            it('should create the provider dependent raw orbiter.yml file', () => {
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

            it('should migrate the api', () => {
                cy.exec(`${Cypress.env('orbctlGitops')} api`)
            })

            it('should write provider secrets', function() {
                Object.entries(this.e2e.initsecrets).forEach((secretRef) => {
                    cy.task('env', secretRef[1]).then(value => {
                        cy.exec(`${Cypress.env('orbctlGitops')} writesecret orbiter.providerundertest.${secretRef[0]}.encrypted --value '${value}'`)
                    })
                })
            });

            execLongRunning(describe, 'should generate some secrets', `${Cypress.env('orbctlGitops')} configure`, 'orbctl', { timeout: 2 * 60 * 1000 })
        })
    })
});

describe('orbctl tests', () => {
   
    it('should not fail when help is passed', () => {
        cy.exec(`${Cypress.env('orbctl')} help`);
    });

    it('should write a secret', () => {
        cy.exec(`${Cypress.env('orbctlGitops')} writesecret boom.monitoring.admin.password.encrypted --value dummy`);
    });

    it('should read a secret', () => {
        cy.exec(`${Cypress.env('orbctlGitops')} readsecret boom.monitoring.admin.password.encrypted`).its('stdout').should('equal', 'dummy')
    })
});

execLongRunning(describe, 'should successfully return when the cluster is bootstrapped and ORBITER took off', `${Cypress.env('orbctlGitops')} takeoff`, 'orbctl', { timeout: 15 * 60 * 1000 })

scale(describe.only, "workers.0", 3, 4, 15 * 60)
scale(describe, "workers.0", 1, 2, 5 * 60)
scale(describe, "controlplane", 3, 4, 15 * 60)
scale(describe, "controlplane", 1, 2, 5 * 60)

function scale(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, pool: string, to: number, expectTotal: number, ensureTimeoutSeconds: number) {

    describe(`init scaling ${pool}`, () => {
        it(`should patch the desired scale of ${pool} to ${to}`, () => {
            cy.exec(`${Cypress.env('orbctlGitops')} file patch orbiter.yml clusters.k8s.spec.${pool}.nodes --exact --value ${to}`)
        })
    })

    describe(`scaling ${pool}`, { retries: Math.ceil(ensureTimeoutSeconds / Cypress.env('retrySeconds')) }, () => {
        it(`should update the infrastructure and the current state to have ${expectTotal} machines`, function() {
            cy.wrap(<Watch>{ operator: 'orbiter', testStart: new Date().toISOString() }).as('watch').exec(`${Cypress.env('orbctlGitops')} file print caos-internal/orbiter/current.yml`).its('stdout').toObject().its('clusters.k8s.current').then(current => {
                expect(Object.keys(current.machines).length, `${pool} machines`).to.equal(expectTotal)
                expect(current.status, 'cluster status').to.equal("running")
                // TODO check nodes with kubectl
            })
        })
    })
}

execLongRunning(describe, 'should destroy the orb', `yes | ${Cypress.env('orbctlGitops')} destroy`, 'orbctl', { timeout: 5 * 60 * 1000})

interface Watch {
    operator: string
    testStart: string
}

function stopAndReport() {

    const ctx = this

    if (!ctx.currentTest.isFailed() || ctx.currentTest.ctx.isLongRunning) {
        cy.wrap(undefined).as("retryMeta").wrap(undefined).as('sinceLatestTrial')
        return
    }

    const totalRetries = ctx.currentTest.retries()
    if (!totalRetries) {
        cy.panic(ctx.currentTest.err)
    }

    const remaining: number = ctx.remainingRetries === undefined ? totalRetries : ctx.remainingRetries
    cy.wrap(remaining - 1).as("remainingRetries").then(remaining => {
        cy.task('info', <LogEntry>{msg: `retrying ${remaining}, totally ${totalRetries} times`, origin: 'cypress'}).then(() => {
            cy.exec('hodor').then(result => cy.task('info', <LogEntry>{msg: result.stdout, origin: 'debug'}))
            /*            
            cy.exec(`kubectl -n caos-system logs --selector "app.kubernetes.io/name=${ctx.currentTest.ctx.watch.operator}" --since-time ${ctx.sinceLatestTrial || ctx.currentTest.ctx.watch.testStart}`).then(result => {
                expect(result.code).to.equal(0)
                return result.stdout.split("\n").forEach((entry) => {
                    var level = 'info'
                    if (entry.indexOf(" err=") > -1) {
                        level = 'error'
                    }
                    cy.task(level, <LogEntry>{msg: entry, origin: ctx.currentTest.ctx.operator})
                })
            }).wrap(ctx.currentTest.watch.testStart).as('sinceLatestTrial')*/
        })
        if (remaining > 0) {
            cy.wait(Cypress.env('retrySeconds') * 1000)
            return
        }
        cy.panic(ctx.currentTest.err)
    })
}

function execLongRunning(describe: Mocha.SuiteFunction | Mocha.ExclusiveSuiteFunction | Mocha.PendingSuiteFunction, itMsg: string, command: string, pkill?: string, options?: Partial<Cypress.ExecOptions>, then?: (chainable: Cypress.Chainable<Cypress.Exec>) => void) {

    describe(`setup command: ${command}`, () => {
        after(function(){
           if (this.currentTest.isFailed() && this.currentTest.err.message.indexOf('timed out after waiting ') > -1) {
               cy.exec(pkill ? `pkill ${pkill}`: 'true', { failOnNonZeroExit: false }).then(() =>
                   cy.readFile(`${Cypress.env('workFolder')}/longrunning.log`).then(log =>
                    cy.panic(`Command timed out.
${log}`
                            )
                    )
                )
            }
        })
       
        describe(`execute command: ${command}`, () => {
            it(itMsg, () => {
                const c = cy.wrap(true).as('isLongRunning').exec(`${command} > ${Cypress.env('workFolder')}/longrunning.log`, options)
                if (then) {
                    then(c)
                }
                c.wrap(false).as('isLongRunning')
            })
        })
    })
}

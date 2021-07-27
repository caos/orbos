describe('install orbctl', { execTimeout: 90000 }, () => {
    // prepare orbctl, download and configuration
    it('download orbctl', () => {
        cy.exec('bash -c \"if [ -f "./orbctl" ]; then rm ./orbctl ; fi\"').its('code').should('eq', 0)
        cy.exec('curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ./orbctl ').its('code').should('eq', 0)
        cy.exec('ls ./orbctl').its('code').should('eq', 0)
    })
    it('chmod orbctl', () => {
        cy.exec('chmod +x ./orbctl ').its('code').should('eq', 0)
    })
})

describe('orbctl configure', { execTimeout: 90000 }, () => {
    it('orbctl configure', () => {
        cy.exec('$orbCtl --gitops --orbconfig $orbConfig --disable-analytics configure --repourl $repoUrl --masterkey "$(openssl rand -base64 21)"', { setTimeout: 20000, env: { orbCtl: Cypress.env("orbCtl"), orbConfig: Cypress.env("orbConfig"), repoUrl: Cypress.env("repoUrl") } }).then(result => {
            cy.log(result.stdout)
            cy.log(result.stderr)
        }).its('code').should('eq', 0)
    })
})



describe('orbctl tests', { execTimeout: 90000 }, () => {
    // basic orbctl operations
    it('orbctl help', () => {
        cy.exec('$orbCtl --gitops --orbconfig $orbConfig --disable-analytics  help ', { setTimeout: 20000, env: { orbCtl: Cypress.env("orbCtl"), orbConfig: Cypress.env("orbConfig") } }).then(result => {
            //return JSON.parse(result.stdout)
            cy.log(result.stdout)
            cy.log(result.stderr)
            //return String.parse(result.stdout)
        }).its('code').should('eq', 0)
    })
   
    it('orbctl writesecret', () => {
        cy.exec('$orbCtl --gitops --orbconfig $orbConfig --disable-analytics  writesecret boom.monitoring.admin.password.encrypted --value $testPassword', { setTimeout: 20000, env: { orbCtl: Cypress.env("orbCtl"), orbConfig: Cypress.env("orbConfig"), testPassword: Cypress.env("testPassword")} }).then(result => {
            cy.log(result.stdout)
            cy.log(result.stderr)
        }).its('code').should('eq', 0)
    })

    it('orbctl readsecret', () => {
        cy.exec('$orbCtl --gitops --orbconfig $orbConfig --disable-analytics  readsecret boom.monitoring.admin.password.encrypted', { setTimeout: 20000, env: { orbCtl: Cypress.env("orbCtl"), orbConfig: Cypress.env("orbConfig"), testPassword: Cypress.env("testPassword") } }).then(result => {
            return result.stdout
        }).then(returnpw => { expect(returnpw).to.eq(Cypress.env("testPassword")) })
    })
})

// describe('orbctl takeoff', { execTimeout: 900000 }, () => {
//     // orbctl takeoff
//     it('orbctl --gitops takeoff', () => {
//         cy.exec('./orbctl takeoff --orbconfig "$CYPRESS_orbConfig" --gitops').its('code').should('eq', 0)
//     })
// })
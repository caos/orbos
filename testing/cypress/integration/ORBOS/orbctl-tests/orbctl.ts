const orbctlGitops = `${Cypress.env("orbCtl")} --gitops --orbconfig ${Cypress.env("orbConfig")} --disable-analytics`

describe('install orbctl', { execTimeout: 90000 }, () => {
    // prepare orbctl, download and configuration
    it('download orbctl', () => {
        cy.exec('bash -c \"if [ -f "./orbctl" ]; then rm ./orbctl ; fi\"').its('code').should('eq', 0);
        cy.exec(`curl -s ${Cypress.env("releaseVersion")} | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ./orbctl `).its('code').should('eq', 0);
        cy.exec('[ -s ./orbctl ]').its('code').should('eq', 0);
        //TODO: check filesize > 0
    });
    it('chmod orbctl', () => {
        cy.exec('chmod +x ./orbctl ').its('code').should('eq', 0);
    })
});

describe('orbctl configure', { execTimeout: 90000 }, () => {
    it('orbctl configure', () => {
        cy.exec(`${orbctlGitops} configure --repourl ${Cypress.env("repoUrl")} --masterkey "$(openssl rand -base64 21)"`, { timeout: 20000 }).then(result => {
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

describe('install orbctl', { execTimeout: 90000 }, () => {
    // prepare orbctl, download and configuration
    it('download orbctl', () => {
        //cy.exec('if [ -f "./orbctl" ]; then rm ./orbctl fi').its('code').should('eq', 0)
        cy.exec('curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d \'\"\' -f 4 | wget -i - -O ./orbctl ').its('code').should('eq', 0)
        cy.exec('ls ./orbctl').its('code').should('eq', 0)
    })
    it('chmod orbctl', () => {
        cy.exec('chmod +x ./orbctl ').its('code').should('eq', 0)
    })
    it('create orbfile', () => {
        //cy.exec('if [[ -f "~/.orb/config" ]]; then rm ~/.orb/config fi').its('code').should('eq', 0)
        cy.exec('./orbctl --gitops configure --repourl "git@github.com:caos/cypress-ops.git" --masterkey "$(openssl rand -base64 21)"').its('code').should('eq', 0)
        cy.exec('ls ~/.orb/config').its('code').should('eq', 0)
    })
})  

describe('orbctl tests', { execTimeout: 90000 }, () => {
    // basic orbctl operations
    it('orbctl help', () => {
        cy.exec('./orbctl help').its('code').should('eq', 0)
    })
    it('orbctl read gpg key from git repository', () => {
        cy.exec('./orbctl readsecret --gitops boom.argocd.gopass.secrets-demo.gpg.encrypted').its('code').should('eq', 0)
    })
})

describe('orbctl takeoff', { execTimeout: 900000 }, () => {
    // orbctl takeoff
    it('orbctl --gitops takeoff', () => {
        cy.exec('./orbctl takeoff --gitops').its('code').should('eq', 0)
    })
})
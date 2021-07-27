Cypress.Commands.add('orbctl',
    (parameter1, parameter2, parameter3, parameter4, parameter5) => {

        cy.exec('$orbCtl --gitops --orbconfig $orbConfig --disable-analytics $parameter1 $parameter2 $parameter3 $parameter4 $parameter5', { setTimeout: 20000, env: { orbCtl: Cypress.env("orbCtl"), orbConfig: Cypress.env("orbConfig"), repoUrl: Cypress.env("repoUrl"), parameter1: parameter1, parameter2: parameter2, parameter3: parameter3, parameter4: parameter4, parameter5: parameter5 } }).then(result => {
            //return JSON.parse(result.stdout)
            cy.log(result.stdout)
            cy.log(result.stderr)
            //return String.parse(result.stdout)
        })
    })

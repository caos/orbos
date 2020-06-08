# Examples to use SSO

This yaml-parts are only examples and there are alot of additional configurations possible, but they should desplay the most used cases.

First-of, all application have to have an DNS-record which can be defined as followed, as an example with grafana:

```yaml
apiVersion: boom.caos.ch/v1beta1
kind: Toolset
metadata:
  name: caos
  namespace: caos-system
spec:
  grafana:
    deploy: true
    network:
      domain: grafana.example.caos.ch
      email: "hi@caos.ch"
      acmeAuthority: "https://acme-staging-v02.api.letsencrypt.org/directory"
```

the same for argocd:
```yaml
  argocd:
    deploy: true
    network:
      domain: argocd.example.caos.ch
      email: "hi@caos.ch"
      acmeAuthority: "https://acme-staging-v02.api.letsencrypt.org/directory"
```

# Grafana

In the IDP used for auth there has to be a registered client with clientID and clientSecret, whereas there also has to be a registered redirectURI. This redirectURI should be *domain-for-grafana*/login/*id*, for example with google: "https://grafana.example.caos.ch/login/google".

All configuration for SSO is under the "auth"-attribute, whereas the domain has to be set correctly so that the redirect works correctly:

```yaml  
apiVersion: boom.caos.ch/v1beta1
kind: Toolset
metadata:
  name: caos
  namespace: caos-system
spec:
  grafana:
    deploy: true
    auth:
```

## Google

The use google as IDP there is the possbility to limit the allowed hosted-domains:

```yaml
      google:
        existingClientIDSecret: 
          name: google-auth
          key: client_id
        existingClientSecretSecret: 
          name: google-auth
          key: client_secret
        allowedDomains:
        - caos.ch
```

## Gitlab

The use google as IDP there is the possbility to limit the allowed groups:

```yaml
      gitlab:
        existingClientIDSecret: 
          name: gitlab-auth
          key: client_id
        existingClientSecretSecret: 
          name: gitlab-auth
          key: client_secret
        allowedGroups:
        - caos
```

## Github

The use google as IDP there is the possbility to limit the allowed organizations:

```yaml
      github:
        existingClientIDSecret: 
          name: github-auth
          key: client_id
        existingClientSecretSecret: 
          name: github-auth
          key: client_secret
        allowedOrganizations:
        - caos
```

## GenericOIDC

To use any generic OIDC as IDP:

```yaml
      genericOAuth:
        existingClientIDSecret: 
          name: oidc-auth
          key: client_id
        existingClientSecretSecret: 
          name: oidc-auth
          key: client_secret
        scopes:
          - openid
          - profile
          - email
        authURL:
        tokenURL:
        apiURL:
        allowedDomains:
          - mycompany.com 
          - mycompany.org
```

# Argocd

In the IDP used for auth there has to be a registered client with clientID and clientSecret, whereas there also has to be a registered redirectURI. This redirectURI should be *domain-for-argocd*/login/*id*, for example with google: "https://argocd.example.caos.ch/api/dex/callback".

All configuration for SSO is under the "auth"-attribute:

```yaml  
apiVersion: boom.caos.ch/v1beta1
kind: Toolset
metadata:
  name: caos
  namespace: caos-system
spec:
  Argocd:
    deploy: true
    auth:
```

## Google

The use google as IDP there is the possbility to limit the allowed hosted-domains:

```yaml
      google:
        id: google
        name: google
        config:
          existingClientIDSecret:
            name: google-auth
            key: client_id
          existingClientSecretSecret:
            name: google-auth
            key: client_secret
          hostedDomains:
          - caos.ch
```

## Gitlab

The use google as IDP there is the possbility to limit the allowed groups:

```yaml
      gitlab:
        id: gitlab
        name: gitlab
        config:
          existingClientIDSecret:
            name: gitlab-auth
            key: client_id
          existingClientSecretSecret:
            name: gitlab-auth
            key: client_secret
          groups:
          - caos
```

## Github

The use google as IDP there is the possbility to limit the allowed organizations:

```yaml
      github:
        id: github
        name: github
        config:
          existingClientIDSecret:
            name: github-auth
            key: client_id
          existingClientSecretSecret:
            name: github-auth
            key: client_secret
          orgs:
          - name: caos
```


## OIDC

To use any generic OIDC as IDP:

```yaml
      oidc:
        Name: unique
        Issuer: test
        existingClientIDSecret:
          name: oidc-auth
          key: client_id
        existingClientSecretSecret:
          name: oidc-auth
          key: client_secret
        RequestedScopes:
          - openid
          - profile
          - email
      # optional
        RequestedIDTokenClaims:
          groups: 
            essential: true
```
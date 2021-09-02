# Example of running the test suite in watch mode

ORBOS_E2E_TAG=cypress-testing-dev ORBOS_E2E_ORBCONFIG=~/.orb/e2egce ORBOS_E2E_ORBURL='git@github.com:caos/ORBOS-Test-GCEProvider.git' ORBOS_E2E_GITHUB_ACCESS_TOKEN="$(gopass show elio-secrets/CAOS/E2E_TOKEN)" ORBOS_E2E_GOOGLE_CLOUD_JSONKEY="$(gopass show caos-secrets/technical/orbos/e2e/gceprovider/jsonkey_base64 | base64 -d)" ginkgo watch --failFast

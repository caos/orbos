# Running e2e tests

Manually triggering an e2e-test:
```bash
go run ./cmd/chore/e2e/trigger/*.go --organization caos --repository ORBOS-Test-StaticProvider --access-token $GITHUB_ACCESS_TOKEN --cleanup=true
go run ./cmd/chore/e2e/trigger/*.go --organization caos --repository ORBOS-Test-GCEProvider --access-token $GITHUB_ACCESS_TOKEN --cleanup=true
```

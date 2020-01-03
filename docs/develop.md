# Develop

Configure your tooling to use certain environment variables. E.g. in VSCode, add the following to your settings.json.

```json
{
    "go.testEnvVars": {
        "MODE": "DEBUG",
        "ORBITER_ROOT": "/home/elio/Code/src/github.com/caos/orbiter"
    },
    "go.testTimeout": "40m",
}
```

Run the tests you find in internal/kinds/clusters/kubernetes/test/kubernetes_test.go in debug mode

For debugging node agents, use a configuration similar to the following VSCode launch.json, adjusting the host IP

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "nodeagent",
            "type": "go",
            "request": "attach",
            "apiVersion": 2,
            "mode": "remote",
            "port": 5000,
            "host": "10.61.0.127"
        },
    ]
}
```

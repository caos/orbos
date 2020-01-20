# Develop

1. Commit your local changes using Commitizen
1. Run a local Orbiter in debug mode by invoking `./scripts/debug.sh ~/.orb/config`
1. In VSCode, use the following launch.json configuration and start a debug session

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
            "host": "127.0.0.1"
        },
    ]
}
```

# Architecture

See [](docs/kind.md)
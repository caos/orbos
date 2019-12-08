module.exports = {
    branch: 'master',
    plugins: [ 
        "@semantic-release/commit-analyzer",
        "@semantic-release/release-notes-generator",
        ["@semantic-release/github", {
            "assets": [
              {"path": "./artifacts/orbctl-darwin-386", "label": "Darwin 386"},
              {"path": "./artifacts/orbctl-darwin-amd64", "label": "Darwin amd64"},
              {"path": "./artifacts/orbctl-freebsd-386", "label": "FreeBSD 386"},
              {"path": "./artifacts/orbctl-freebsd-amd64", "label": "FreeBSD amd64"},
              {"path": "./artifacts/orbctl-linux-386", "label": "Linux 386"},
              {"path": "./artifacts/orbctl-linux-amd64", "label": "Linux amd64"},
              {"path": "./artifacts/orbctl-openbsd-386", "label": "OpenBSD 386"},
              {"path": "./artifacts/orbctl-openbsd-amd64", "label": "OpenBSD amd64"},
              {"path": "./artifacts/orbctl-windows-386", "label": "Windows 386"},
              {"path": "./artifacts/orbctl-windows-amd64", "label": "Windows amd64"}
            ]
        }]
    ]
  };
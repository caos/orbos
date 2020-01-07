# Troubleshooting

If you find yourself in a problem with `Orbiter` we hopefully provide you some helpfull hints in the list below.

## Problems

### Orbiter is stuck in state `container creating`

- Check if the `kubernetes` network is up and running

### Orbiter cannot connect to the node by SSH

If you see this error {CANNOT_CONNECT} you want to make sure that:

- `Orbiter` has access by SSH to all the Servers on their HostIP
- With cloud providers `Orbiter` uses the cloud provider specific `SSH` implementation
  - For example with Google Cloud it uses `Google Cloud Console`

If you see this error {WRONGKEY} you want to make sure that:

- By default `Orbiter` connects to a server by utilzining a key called `bootstrap key` after the initial connection this one will be replaced by:
  - ... freshly generated keypair
  - ... preexisiting keypair in the Git Repo
  - with this process each cluster get's it's own set of keys

### Orbiter cannot connect to the kube API

If you see this error {CANNOT_CONNECT} you want to make sure that:

- Has access to the loadbalancers holding the virtual ip address of the kubeapi
  - By default this should be the port 6443
  - The Loadbalancer creates a bind to {kube-api-ip}:6443 and forwads traffic to {node-ip}:6666
  - It is not recommended to connect to the port 6666 directly

### Orbiter cannot connect to the Git Repository

If you see this error {CANNOT_CONNECT} you want to make sure that:

- `Orbiter` has SSH Connectivity to you Git Server

If you see this error {WRONGKEY_SSH_Key} you want to make sure that:

- That the secret mounted to the `Orbiter` deployment contains the correct SSH key,
  - This key is used to connect by `SSH` to the Repo

If you see this error {WRONGKEY_MASTER_Key} you want to make sure that:

- That the secret mounted to the `Orbiter` deployment contains the correct Master key,
  - This key is used to decrypt the `secrect` provided with the Repo

# Setup an Orb

An Orb is the combination of all layers between the provider(for example Google Cloud Engine or on-premise VMs) and Kubernetes inclusive all tool provided by Orbos.
The configuration consists of a git-repository with all desired states and the orb file.

## Initialize A Git Repository

Generate a new Deploy Key
```bash
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "ORBOS repo key" -P "" -f /tmp/myorb_repo -q
```

Create a new Git Repository

Add the public part of your new SSH key pair to the git repositories trusted deploy keys with write access.

```
cat /tmp/myorb_repo.pub
```

Copy the files [orbiter.yml](../../examples/orbiter/gce/orbiter.yml) and [boom.yml](../../examples/boom/boom.yml) to the root of your Repository.

## Configure your local environment

Download the latest orbctl

```bash
curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl


# Create an orb file at ${HOME}/.orb/config
orbctl configure --repourl git@github.com:me/my-orb.git --masterkey "$(openssl rand -base64 21)"
```
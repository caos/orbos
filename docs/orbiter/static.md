# Using the StaticProvider

In the following example we will create a `kubernetes` cluster on a `StaticProvider`. A `StaticProvider` is a provider, which has no or little API for automation, e.g legacy VM's or Bare Metal scenarios. For demonstration purposes, we use KVM to provision VM's here.

## Create Two Virtual Machines

Install KVM according to the [docs](https://wiki.debian.org/KVM)

Create a new SSH key pair

```bash
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "repo and VM bootstrap key" -P "" -f ~/.ssh/myorb_bootstrap -q
```

Create and setup two new Virtual Machines. Make sure you have a sudo user called orbiter on the guest OS

```bash
./examples/k8s/static/machine.sh ./examples/k8s/static/kickstart.cfg ~/.ssh/myorb_bootstrap.pub master1
./examples/k8s/static/machine.sh ./examples/k8s/static/kickstart.cfg ~/.ssh/myorb_bootstrap.pub worker1
```

List the new virtual machines IP addresses

```bash
for MACHINE in master1 worker1
do
    virsh domifaddr $MACHINE
done
```

## Initialize A Git Repository

Create a new Git Repository. Add the public part of your new SSH key pair to the git repositories trusted deploy keys with write access.

```
cat ~/.ssh/myorb_bootstrap.pub
```

Copy the files [orbiter.yml](../../examples/orbiter/static/orbiter.yml) and [boom.yml](../../examples/boom/boom.yml) to the root of your Repository.

Replace the IPs in your orbiter.yml accordingly

## Complete Your Orb Setup

Download the latest orbctl

```bash
curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl
```

Create an orb file

```bash
mkdir -p ~/.orb
cat > ~/.orb/config << EOF
url: git@github.com:me/my-orb.git
masterkey: $(openssl rand -base64 21)
repokey: |
$(sed s/^/\ \ /g ~/.ssh/myorb_bootstrap)
EOF
```

Encrypt and write your ssh key pair to your repo

```bash
# Add the bootstrap key pair to the remote secrets file. For simplicity, we use the repokey here.
orbctl writesecret kvm.bootstrapkeyprivate --file ~/.ssh/myorb_bootstrap
orbctl writesecret kvm.bootstrapkeypublic --file ~/.ssh/myorb_bootstrap.pub
```
## Bootstrap your local Kubernetes cluster

```bash
orbctl takeoff
```

As soon as the Orbiter has deployed itself to the cluster, you can decrypt the generated admin kubeconfig

```bash
mkdir -p ~/.kube
orbctl readsecret k8s.kubeconfig > ~/.kube/config
```

Wait for grafana to become running

```bash
kubectl --namespace caos-system get po -w
```

Open your browser at localhost:8080 to show your new clusters dashboards

```bash
kubectl --namespace caos-system port-forward svc/grafana 8080:80
```

Cleanup your environment

```bash
for MACHINE in master1 worker1
do
    virsh destroy $MACHINE
    virsh undefine $MACHINE
done
```

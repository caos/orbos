# Using the StaticProvider

In the following example we will create a `kubernetes` cluster on a `StaticProvider`. A `StaticProvider` is a provider, which has no or little API for automation, e.g legacy VM's or Bare Metal scenarios. For demonstration purposes, we use KVM to provision VM's here.

## Create Two Virtual Machines

Install KVM according to the [docs](https://wiki.debian.org/KVM)

Create a new SSH key pair

```bash
mkdir -p ~/.ssh && ssh-keygen -t rsa -b 4096 -C "VM bootstrap key" -P "" -f ~/.ssh/myorb_bootstrap -q
```

Create and setup two new Virtual Machines. Make sure you have a sudo user called orbiter on the guest OS

```bash
./examples/orbiter/static/kvm/machine.sh ./examples/orbiter/static/kvm/kickstart.cfg ~/.ssh/myorb_bootstrap.pub master1
./examples/orbiter/static/kvm/machine.sh ./examples/orbiter/static/kvm/kickstart.cfg ~/.ssh/myorb_bootstrap.pub worker1
```

### Create a new repository on Github.com

Copy the files [orbiter.yml](examples/orbiter/gce/orbiter.yml) and [boom.yml](examples/boom/boom.yml) to the root of your new git Repository

### Configure your environment
```bash
# Install the latest orbctl
curl -s https://api.github.com/repos/caos/orbos/releases/latest | grep "browser_download_url.*orbctl-$(uname)-$(uname -m)" | cut -d '"' -f 4 | sudo wget -i - -O /usr/local/bin/orbctl
sudo chmod +x /usr/local/bin/orbctl
sudo chown $(id -u):$(id -g) /usr/local/bin/orbctl

# Create an orb file at ${HOME}/.orb/config
orbctl configure --repourl 'git@github.com:me/my-orb.git' --masterkey "$(openssl rand -base64 21)"

# Add your bootstrap key pair to the remote orbiter.yml
orbctl writesecret orbiter.kvm.bootstrapkeyprivate --file ~/.ssh/myorb_bootstrap
orbctl writesecret orbiter.kvm.bootstrapkeypublic --file ~/.ssh/myorb_bootstrap.pub

# Note your machine names and IP addresses
for VM in $(virsh list --all --name); do echo $VM; virsh domifaddr $VM; done

# Update the pools section according to the output of the following command and push your changes to the remote repository
orbctl edit orbiter.yml
```

## Bootstrap your local Kubernetes cluster

```bash
orbctl takeoff
```

As soon as the Orbiter has deployed itself to the cluster, you can decrypt the generated admin kubeconfig

```bash
mkdir -p ~/.kube
orbctl readsecret orbiter.k8s.kubeconfig > ~/.kube/config
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

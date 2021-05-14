# CA-KeyPair Operator

> **ATTENTION:** THE `MASTER` BRANCH MAY BE IN AN UNSTABLE OR EVEN BROKEN STATE DURING DEVELOPMENT.

[cert-manager](https://cert-manager.io/) is a very fine solution if you want to run you own CA for signing certificates.

In order for the cert-manager to sign certificates however, there needs to be a CA key pair for the cert-manager first.

The CA-KeyPair Operator allows you to create CA key pairs in Kubernetes.

## Getting started

### Install CA-KeyPair Operator with helm
You can install the CA-KeyPair Operator with [Helm 3](https://helm.sh/).

```shell
# Add the helm repo
$ helm repo add cakeypair-operator https://raw.githubusercontent.com/dajudge/cakeypair-operator/master/charts/

# Create a namespace for the operator
$ kubectl create ns cakeypair-operator

# Install the latest operator version 
$ helm install -n cakeypair-operator cakeypair-operator cakeypair-operator/cakeypair-operator
```

### Deploy a key pair
Deploy a `CaKeyPair` custom resource:

```yaml
apiVersion: cakeypairs.dajudge.com/v1alpha1
kind: CaKeyPair
metadata:
  name: cakeypair-sample
spec:
  keySize: 4096
  secretName: ca-key-pair
  commonName: My own root CA
  subject:
    organizations: [ "Acme Corp." ]
    organizationalUnits: [ "Cyber-Security Global Operations", "CS-GO" ]
    countries: [ "Lampukistan" ]
```

## Additional references

* [cert-manager](https://cert-manager.io/docs/)
* [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
* [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Kubebuilder](https://book.kubebuilder.io/)

## License

See [LICENSE](https://github.com/dajudge/cakeypair-operator/blob/master/LICENSE) for licensing details.
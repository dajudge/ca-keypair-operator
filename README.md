CA-KeyPair Operator
-
> **ATTENTION:** THE `MASTER` BRANCH MAY BE IN AN UNSTABLE OR EVEN BROKEN STATE DURING DEVELOPMENT.

# TL;DR
The CA-KeyPair Operator allows you to easily initialize random CA key pairs in Kubernetes.

# Description
[cert-manager](https://cert-manager.io/) is (among other things) a very fine solution if you want to run you own
CA in Kubernetes for signing certificates.

In order for the cert-manager to sign certificates however, there needs to be a CA key pair for the cert-manager to pick
up, first. This can be daunting to put in place if you're only interested in a randomly initialized CA.

The CA-KeyPair operator introduces a CRD to describe the key pairs you want to initialize and creates random key pairs
in Kubernetes secrets based on these specifications.

# Getting started

## Install CA-KeyPair Operator with helm
You can install the CA-KeyPair Operator with [Helm 3](https://helm.sh/).

```shell
# Add the helm repo
$ helm repo add cakeypair-operator https://raw.githubusercontent.com/dajudge/cakeypair-operator/master/charts/

# Create a namespace for the operator
$ kubectl create ns cakeypair-operator

# Install the latest operator version 
$ helm install -n cakeypair-operator cakeypair-operator cakeypair-operator/cakeypair-operator
```

## Deploy a key pair
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

# Additional references

* [cert-manager](https://cert-manager.io/docs/)
* [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
* [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Kubebuilder](https://book.kubebuilder.io/)

# License

See [LICENSE](https://github.com/dajudge/cakeypair-operator/blob/master/LICENSE) for licensing details.

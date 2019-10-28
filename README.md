# spire-k8s-operator

This operator provides CRDs for SpiffeIDs.

At the moment it supports a non-namespaced ClusterSpiffeID, with no restrictions on the spiffe IDs it can create.

It also optionally provides a controller that emulates the older k8s-registrar behaviour, creating and destroying SpiffeId resources based on Pods.

It is a very early work in progress.

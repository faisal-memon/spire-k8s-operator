package controller

import (
	"github.com/transferwise/spire-k8s-operator/pkg/controller/spiffeid"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, spiffeid.Add)
}

package controller

import (
	"github.com/spiffe/spire/proto/spire/api/registration"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, registration.RegistrationClient) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, r registration.RegistrationClient) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, r); err != nil {
			return err
		}
	}
	return nil
}

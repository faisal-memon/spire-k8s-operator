package spiremgr

import (
	"context"
	"github.com/go-logr/logr"
	spiffeidv1alpha1 "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Finalizer struct {
	Client      client.Client
	FinalizerName string
}

func (r *Finalizer) Finalizable(instance v1.Object) bool {
    return instance.GetDeletionTimestamp() != nil
}

func (r *Finalizer) Finalize(reqLogger logr.Logger, instance spiffeidv1alpha1.CommonSpiffeId, finalizer func() error) error {
		if contains(instance.GetFinalizers(), r.FinalizerName) {
			reqLogger.Info("Finalizing...")
			// Run finalization logic. If the finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := finalizer(); err != nil {
				reqLogger.Error(err, "Failed to finalize ID")
				return err
			}

			// Remove spiffeIdFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			reqLogger.Info("Finalized")
			instance.SetFinalizers(remove(instance.GetFinalizers(), r.FinalizerName))
			err := r.Client.Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to mark instance finalized")
				return err
			}
		}
		return nil
}

func (r *Finalizer) AddFinalizer(reqLogger logr.Logger, instance spiffeidv1alpha1.CommonSpiffeId) error {
	if !contains(instance.GetFinalizers(), r.FinalizerName) {
		reqLogger.Info("Adding Finalizer for SpiffeId")
		instance.SetFinalizers(append(instance.GetFinalizers(), r.FinalizerName))

		// Update CR
		err := r.Client.Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update SpiffeId with finalizer")
			return err
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, v := range list {
		if v == s {
			list = append(list[:i], list[i+1:]...)
		}
	}
	return list
}

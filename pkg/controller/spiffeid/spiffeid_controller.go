package spiffeid

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/spiffe/spire/proto/spire/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spiffe/spire/proto/spire/api/registration"
	spiffeidv1alpha1 "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const spiffeIdFinalizer = "finalizer.spiffeid.spiffe.io"

var log = logf.Log.WithName("controller_spiffeid")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SpiffeId Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSpiffeId{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("spiffeid-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SpiffeId
	err = c.Watch(&source.Kind{Type: &spiffeidv1alpha1.SpiffeId{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSpiffeId implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSpiffeId{}

// ReconcileSpiffeId reconciles a SpiffeId object
type ReconcileSpiffeId struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	spireClient registration.RegistrationClient
}

// Reconcile reads that state of the cluster for a SpiffeId object and makes changes based on the state read
// and what is in the SpiffeId.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSpiffeId) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SpiffeId")

	// Fetch the SpiffeId instance
	instance := &spiffeidv1alpha1.SpiffeId{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// If the resource is not found, that means all of
			// the finalizers have been removed, and the memcached
			// resource has been deleted, so there is nothing left
			// to do.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil {
		if !contains(instance.GetFinalizers(), spiffeIdFinalizer) {
			// Run finalization logic. If the finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeSpiffeId(reqLogger, instance); err != nil {
				return reconcile.Result{}, err
			}

			// Remove spiffeIdFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			instance.SetFinalizers(remove(instance.GetFinalizers(), spiffeIdFinalizer))
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	if !contains(instance.GetFinalizers(), spiffeIdFinalizer) {
		if err := r.addFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	entryId, err := r.createSpireEntry(reqLogger, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	instance.Status.EntryId = entryId
	err = r.client.Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Failed to update SpiffeId with Entry ID" + entryId)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSpiffeId) createSpireEntry(reqLogger logr.Logger, instance *spiffeidv1alpha1.SpiffeId) (string, error) {
	entryId, err := r.spireClient.CreateEntry(context.TODO(), &common.RegistrationEntry{
		Selectors: nil,
		ParentId:  "",
		SpiffeId:  instance.Spec.SpiffeId,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
		  // TODO: return existing entry ID?
		}
		return "", err
	}

	return entryId.Id, nil
}


func (r *ReconcileSpiffeId) finalizeSpiffeId(reqLogger logr.Logger, instance *spiffeidv1alpha1.SpiffeId) error {
	regEntryId := &registration.RegistrationEntryID{
		Id: instance.Status.EntryId,
	}
	_, err := r.spireClient.DeleteEntry(context.TODO(), regEntryId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		reqLogger.Error(err, "Failed to delete registration entry " + regEntryId.Id)
		return err
	}
	reqLogger.Info("Successfully finalized spiffeId")
	return nil
}

func (r *ReconcileSpiffeId) addFinalizer(reqLogger logr.Logger, instance *spiffeidv1alpha1.SpiffeId) error {
	reqLogger.Info("Adding Finalizer for SpiffeId")
	instance.SetFinalizers(append(instance.GetFinalizers(), spiffeIdFinalizer))

	// Update CR
	err := r.client.Update(context.TODO(), instance)
	if err != nil {
		reqLogger.Error(err, "Failed to update SpiffeId with finalizer")
		return err
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

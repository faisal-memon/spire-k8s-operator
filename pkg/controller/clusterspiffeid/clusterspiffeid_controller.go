package clusterspiffeid

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spiffe/spire/proto/spire/api/registration"
	"github.com/spiffe/spire/proto/spire/common"
	spiffeidv1alpha1 "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1"
	"github.com/transferwise/spire-k8s-operator/pkg/spiremgr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const spiffeIdFinalizer = "finalizer.clusterspiffeid.spiffe.io"

var log = logf.Log.WithName("controller_clusterspiffeid")

// Add creates a new SpiffeId Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, r registration.RegistrationClient, conf ReconcileClusterSpiffeIdConfig) error {
	return add(mgr, newReconciler(mgr, r, conf))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, r registration.RegistrationClient, conf ReconcileClusterSpiffeIdConfig) reconcile.Reconciler {
    return &ReconcileClusterSpiffeId{client: mgr.GetClient(), scheme: mgr.GetScheme(), spireClient: r, conf: conf, utils: spiremgr.SpireUtils{SpireClient: r}, finalizer: spiremgr.Finalizer{Client: mgr.GetClient(), FinalizerName: spiffeIdFinalizer}}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterspiffeid-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SpiffeId
	err = c.Watch(&source.Kind{Type: &spiffeidv1alpha1.ClusterSpiffeId{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileClusterSpiffeId implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClusterSpiffeId{}

type ReconcileClusterSpiffeIdConfig struct {
	TrustDomain string
	Cluster     string
}

// ReconcileClusterSpiffeId reconciles a SpiffeId object
type ReconcileClusterSpiffeId struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	spireClient registration.RegistrationClient
	conf        ReconcileClusterSpiffeIdConfig
	utils       spiremgr.SpireUtils
	finalizer   spiremgr.Finalizer
}

// Reconcile reads that state of the cluster for a SpiffeId object and makes changes based on the state read
// and what is in the SpiffeId.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterSpiffeId) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SpiffeId")

	// Fetch the SpiffeId instance
	instance := &spiffeidv1alpha1.ClusterSpiffeId{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if k8errors.IsNotFound(err) {
			// If the resource is not found, that means all of
			// the finalizers have been removed, and the
			// resource has been deleted, so there is nothing left
			// to do.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

    if r.finalizer.Finalizable(instance) {
        if err := r.finalizer.Finalize(log, instance, func() error {
          return r.utils.DeleteEntry(log, instance.Status.EntryId)
        }); err != nil {
            return reconcile.Result{}, err
        }
        return reconcile.Result{}, nil
    }

	if err := r.finalizer.AddFinalizer(reqLogger, instance); err != nil {
			return reconcile.Result{}, err
		}

	entryId, err := r.createSpireEntry(reqLogger, instance)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists && instance.Status.EntryId == entryId {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.Status.EntryId != entryId {
		instance.Status.EntryId = entryId
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update SpiffeId to add entry ID", "entryID", entryId)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterSpiffeId) createSpireEntry(reqLogger logr.Logger, instance *spiffeidv1alpha1.ClusterSpiffeId) (string, error) {

	// TODO: sanitize!
	selectors := make([]*common.Selector, 0, len(instance.Spec.Selector.PodLabel))
	for k, v := range instance.Spec.Selector.PodLabel {
		selectors = append(selectors, &common.Selector{Value: fmt.Sprintf("k8s:pod-label:%s:%s", k, v)})
	}
	if len(instance.Spec.Selector.PodName) > 0 {
		selectors = append(selectors, &common.Selector{Value: fmt.Sprintf("k8s:pod-name:%s", instance.Spec.Selector.PodName)})
	}
	if len(instance.Spec.Selector.Namespace) > 0 {
		selectors = append(selectors, &common.Selector{Value: fmt.Sprintf("k8s:ns:%s", instance.Spec.Selector.Namespace)})
	}
	if len(instance.Spec.Selector.ServiceAccount) > 0 {
		selectors = append(selectors, &common.Selector{Value: fmt.Sprintf("k8s:sa:%s", instance.Spec.Selector.ServiceAccount)})
	}
	for _, v := range instance.Spec.Selector.Arbitrary {
		// TODO: Protect with some flag because this seems like it could be a really bad idea...
		selectors = append(selectors, &common.Selector{Value: v})
	}

	return r.utils.GetOrCreateEntry(reqLogger, instance.Spec.SpiffeId, selectors)
}

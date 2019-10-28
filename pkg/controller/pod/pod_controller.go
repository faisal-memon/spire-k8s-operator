package pod

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"path"

	"github.com/spiffe/spire/proto/spire/api/registration"
	spiffeidv1alpha1 "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_pod")

// Add creates a new SpiffeId Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, conf PodReconcilerConfig) error {
	return add(mgr, newReconciler(mgr, conf))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, conf PodReconcilerConfig) reconcile.Reconciler {
	return &ReconcilePod{client: mgr.GetClient(), scheme: mgr.GetScheme(), config: conf}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Pod
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &spiffeidv1alpha1.ClusterSpiffeId{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.Pod{},
	})

	return nil
}

// blank assignment to verify that ReconcilePod implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePod{}

type PodReconcilerMode int32

const (
	PodReconcilerModeServiceAccount PodReconcilerMode = iota
	PodReconcilerModeLabel
	PodReconcilerModeAnnotation
)

type PodReconcilerConfig struct {
	TrustDomain string
	Mode        PodReconcilerMode
	Value       string
}

// ReconcilePod reconciles a Pod object
type ReconcilePod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	spireClient registration.RegistrationClient
	config      PodReconcilerConfig
}

// Reconcile reads that state of the cluster for a SpiffeId object and makes changes based on the state read
// and what is in the SpiffeId.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcilePod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the Pod instance
	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pod)
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

	spiffeidname := fmt.Sprintf("spire-operator-%s", pod.GetName())

	spiffeId := ""
	switch r.config.Mode {
	case PodReconcilerModeServiceAccount:
		spiffeId = r.makeID("ns/%s/sa/%s", request.Namespace, pod.Spec.ServiceAccountName)
	case PodReconcilerModeLabel:
		if val, ok := pod.GetLabels()[r.config.Value]; ok {
			spiffeId = r.makeID("%s", val)
		} else {
			// No relevant label
			return reconcile.Result{}, nil
		}
	case PodReconcilerModeAnnotation:
		if val, ok := pod.GetAnnotations()[r.config.Value]; ok {
			spiffeId = r.makeID("%s", val)
		} else {
			// No relevant annotation
			return reconcile.Result{}, nil
		}
	}
	reqLogger.Info("Reconciling Pod")

	existing := &spiffeidv1alpha1.ClusterSpiffeId{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: spiffeidname}, existing)
	if err != nil && k8errors.IsNotFound(err) {
		clusterSpiffeId := &spiffeidv1alpha1.ClusterSpiffeId{
			ObjectMeta: v1.ObjectMeta{
				Name: spiffeidname,
			},
			Spec: spiffeidv1alpha1.SpiffeIdSpec{
				SpiffeId: spiffeId,
				Selector: spiffeidv1alpha1.Selector{
					PodName: pod.Name,
				},
			},
		}
		err = controllerutil.SetControllerReference(pod, clusterSpiffeId, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to create new SpiffeID", "SpiffeID.Name", clusterSpiffeId.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new SpiffeID", "SpiffeID.Name", clusterSpiffeId.Name)
		err = r.client.Create(context.TODO(), clusterSpiffeId)
		if err != nil {
			reqLogger.Error(err, "Failed to create new SpiffeID", "SpiffeID.Name", clusterSpiffeId.Name)
			return reconcile.Result{}, err
		}
		// SpiffeID created successfully
		return reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get SpiffeID", "name", spiffeidname)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcilePod) makeID(pathFmt string, pathArgs ...interface{}) string {
	id := url.URL{
		Scheme: "spiffe",
		Host:   r.config.TrustDomain,
		Path:   path.Clean(fmt.Sprintf(pathFmt, pathArgs...)),
	}
	return id.String()
}

package spiffeid

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spiffe/spire/proto/spire/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"path"

	"github.com/spiffe/spire/proto/spire/api/registration"
	spiffeidv1alpha1 "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1"
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

const spiffeIdFinalizer = "finalizer.spiffeid.spiffe.io"

var log = logf.Log.WithName("controller_spiffeid")

// Add creates a new SpiffeId Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, r registration.RegistrationClient, conf ReconcileSpiffeIdConfig) error {
	return add(mgr, newReconciler(mgr, r, conf))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, r registration.RegistrationClient, conf ReconcileSpiffeIdConfig) reconcile.Reconciler {
	return &ReconcileSpiffeId{client: mgr.GetClient(), scheme: mgr.GetScheme(), spireClient: r, conf: conf}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("spiffeid-controller", mgr, controller.Options{Reconciler: r})
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

// blank assignment to verify that ReconcileSpiffeId implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSpiffeId{}

type ReconcileSpiffeIdConfig struct {
	TrustDomain string
	Cluster     string
}

// ReconcileSpiffeId reconciles a SpiffeId object
type ReconcileSpiffeId struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client      client.Client
	scheme      *runtime.Scheme
	spireClient registration.RegistrationClient
	conf		 ReconcileSpiffeIdConfig
	myId        *string
}

// Reconcile reads that state of the cluster for a SpiffeId object and makes changes based on the state read
// and what is in the SpiffeId.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSpiffeId) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SpiffeId")

	if r.myId == nil {
		err := r.setMyId()
		if err != nil {
			return reconcile.Result{}, err
		}
	}

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

	if instance.GetDeletionTimestamp() != nil {
		if contains(instance.GetFinalizers(), spiffeIdFinalizer) {
			log.Info("Finalizing...")
			// Run finalization logic. If the finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			if err := r.finalizeSpiffeId(reqLogger, instance); err != nil {
				log.Error(err, "Failed to finalize ID")
				return reconcile.Result{}, err
			}

			// Remove spiffeIdFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			log.Info("Finalized")
			instance.SetFinalizers(remove(instance.GetFinalizers(), spiffeIdFinalizer))
			err := r.client.Update(context.TODO(), instance)
			if err != nil {
				log.Error(err, "Failed to mark instance finalized")
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

func (r *ReconcileSpiffeId) makeID(pathFmt string, pathArgs ...interface{}) string {
	id := url.URL{
		Scheme: "spiffe",
		Host:   r.conf.TrustDomain,
		Path:   path.Clean(fmt.Sprintf(pathFmt, pathArgs...)),
	}
	return id.String()
}


// ServerID creates a server SPIFFE ID string given a trustDomain.
func ServerID(trustDomain string) string {
	return ServerURI(trustDomain).String()
}

// ServerURI creates a server SPIFFE URI given a trustDomain.
func ServerURI(trustDomain string) *url.URL {
	return &url.URL{
		Scheme: "spiffe",
		Host:   trustDomain,
		Path:   path.Join("spire", "server"),
	}
}


func (r *ReconcileSpiffeId) nodeID() string {
	return r.makeID("spire-k8s-operator/%s/node", r.conf.Cluster)
}

func (r *ReconcileSpiffeId) setMyId() error {
	myId := r.nodeID()
	log.Info("Initializing operator parent ID.")
	_, err := r.spireClient.CreateEntry(context.TODO(), &common.RegistrationEntry{
		Selectors: []*common.Selector{
			{Type: "k8s_psat", Value: fmt.Sprintf("cluster:%s", r.conf.Cluster)},
		},
		ParentId:  ServerID(r.conf.TrustDomain),
		SpiffeId:  myId,
	})
	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			log.Info("Failed to create operator parent ID", "spiffeID", myId)

			return err
		}
	}
	log.Info("Initialized operator parent ID", "spiffeID", myId)
	r.myId = &myId
	return nil
}

var ExistingEntryNotFoundError = errors.New("No existing matching entry found")

func (r *ReconcileSpiffeId) getExistingEntry(reqLogger logr.Logger, id string, selectors []*common.Selector) (string, error) {
	entries, err := r.spireClient.ListByParentID(context.TODO(), &registration.ParentID{
		Id: *r.myId,
	})
	if err != nil {
		reqLogger.Error(err, "Failed to retrieve existing spire entry")
		return "", err
	}

	selectorMap := map[string]map[string]bool{}
	for _, sel := range selectors {
		if _, ok := selectorMap[sel.Type]; !ok {
			selectorMap[sel.Type] = make(map[string]bool)
		}
		selectorMap[sel.Type][sel.Value] = true
	}
	for _, entry := range entries.Entries {
		if entry.GetSpiffeId() == id {
			if len(entry.GetSelectors()) != len(selectors) {
				continue
			}
			for _, sel := range entry.GetSelectors() {
				if _, ok := selectorMap[sel.Type]; !ok {
					continue
				}
				if _, ok := selectorMap[sel.Type][sel.Value]; !ok {
					continue
				}
			}
			return entry.GetEntryId(), nil
		}
	}
	return "", ExistingEntryNotFoundError
}

func (r *ReconcileSpiffeId) createSpireEntry(reqLogger logr.Logger, instance *spiffeidv1alpha1.ClusterSpiffeId) (string, error) {

	// TODO: sanitize!
	selectors := make([]*common.Selector, 0, len(instance.Spec.Selector.PodLabel))
	for k, v := range instance.Spec.Selector.PodLabel {
		selectors = append(selectors, &common.Selector{Value: fmt.Sprintf("k8s:pod-label:%s:%s", k, v)})
	}

	reqLogger.Info("Creating entry", "spiffeID", instance.Spec.SpiffeId)

	regEntryId, err := r.spireClient.CreateEntry(context.TODO(), &common.RegistrationEntry{
		Selectors: selectors,
		ParentId:  *r.myId,
		SpiffeId:  instance.Spec.SpiffeId,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			// TODO: How to prevent accidental deletion if someone creates duplicate entries
			// Is the finalizer actually the right way to clean up?
			// Use another CRD to provide tracking of resources?
			entryId, err := r.getExistingEntry(reqLogger, instance.Spec.SpiffeId, selectors)
			if err != nil {
				reqLogger.Error(err, "Failed to reuse existing spire entry")
				return "", err
			}
			reqLogger.Info("Reused existing entry", "entryID", entryId, "spiffeID", instance.Spec.SpiffeId)
			return entryId, nil
		}
		reqLogger.Error(err, "Failed to create spire entry")
		return "", err
	}
	reqLogger.Info("Created entry", "entryID", regEntryId.Id, "spiffeID", instance.Spec.SpiffeId)

	return regEntryId.Id, nil
}

func (r *ReconcileSpiffeId) finalizeSpiffeId(reqLogger logr.Logger, instance *spiffeidv1alpha1.ClusterSpiffeId) error {
	regEntryId := &registration.RegistrationEntryID{
		Id: instance.Status.EntryId,
	}
	_, err := r.spireClient.DeleteEntry(context.TODO(), regEntryId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		// Spire server returns internal server error rather than NotFound when the entry doesn't exist.
		//reqLogger.Error(err, "Failed to delete registration entry", "entryID", regEntryId.Id)
		//return err
		reqLogger.Error(err,"Got error deleting spire entry, but assuming it's OK")
		return nil
	}
	reqLogger.Info("Successfully finalized spiffeId")
	return nil
}

func (r *ReconcileSpiffeId) addFinalizer(reqLogger logr.Logger, instance *spiffeidv1alpha1.ClusterSpiffeId) error {
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

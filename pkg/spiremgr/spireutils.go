package spiremgr

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/spiffe/spire/proto/spire/api/registration"
	"github.com/spiffe/spire/proto/spire/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/url"
	"path"
)

type SpireUtils struct {
	SpireClient registration.RegistrationClient
	TrustDomain	string
	Cluster	    string
	myId        *string
}


func (r *SpireUtils) makeID(pathFmt string, pathArgs ...interface{}) string {
	id := url.URL{
		Scheme: "spiffe",
		Host:   r.TrustDomain,
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

func (r *SpireUtils) nodeID() string {
	return r.makeID("spire-k8s-operator/%s/node", r.Cluster)
}

func (r *SpireUtils) makeMyId(reqLogger logr.Logger) (string, error) {
	myId := r.nodeID()
	reqLogger.Info("Initializing operator parent ID.")
	_, err := r.SpireClient.CreateEntry(context.TODO(), &common.RegistrationEntry{
		Selectors: []*common.Selector{
			{Type: "k8s_psat", Value: fmt.Sprintf("cluster:%s", r.Cluster)},
		},
		ParentId: ServerID(r.TrustDomain),
		SpiffeId: myId,
	})
	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			reqLogger.Info("Failed to create operator parent ID", "spiffeID", myId)
			return "", err
		}
	}
	reqLogger.Info("Initialized operator parent ID", "spiffeID", myId)
	return myId, nil
}

func (r *SpireUtils) getMyId(reqLogger logr.Logger) (string, error) {
	if r.myId == nil {
		myId, err := r.makeMyId(reqLogger)
		if err != nil {
			return "", err
		}
		r.myId = &myId
	}
	return *r.myId, nil
}

func (r *SpireUtils) DeleteEntry(reqLogger logr.Logger, entryId string) error {
	regEntryId := &registration.RegistrationEntryID{
		Id: entryId,
	}
	_, err := r.SpireClient.DeleteEntry(context.TODO(), regEntryId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil
		}
		// Spire server returns internal server error rather than NotFound when the entry doesn't exist.
		//reqLogger.Error(err, "Failed to delete registration entry", "entryID", regEntryId.Id)
		//return err
		reqLogger.Error(err, "Got error deleting spire entry, but assuming it's OK")
		return nil
	}
	reqLogger.Info("Successfully finalized spiffeId")
	return nil
}

var ExistingEntryNotFoundError = errors.New("No existing matching entry found")

func (r *SpireUtils) getExistingEntry(reqLogger logr.Logger, id string, selectors []*common.Selector) (string, error) {
	myId, err := r.getMyId(reqLogger)
	if err != nil {
		return "", err
	}
	entries, err := r.SpireClient.ListByParentID(context.TODO(), &registration.ParentID{
		Id: myId,
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

func (r *SpireUtils) GetOrCreateEntry(reqLogger logr.Logger, spiffeId string, selectors []*common.Selector) (string, error) {

	reqLogger.Info("Creating entry", "spiffeID", spiffeId)

	myId, err := r.getMyId(reqLogger)
	if err != nil {
		return "", err
	}

	regEntryId, err := r.SpireClient.CreateEntry(context.TODO(), &common.RegistrationEntry{
		Selectors: selectors,
		ParentId:  myId,
		SpiffeId:  spiffeId,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			entryId, err := r.getExistingEntry(reqLogger, spiffeId, selectors)
			if err != nil {
				reqLogger.Error(err, "Failed to reuse existing spire entry")
				return "", err
			}
			reqLogger.Info("Found existing entry", "entryID", entryId, "spiffeID", spiffeId)
			return entryId, err
		}
		reqLogger.Error(err, "Failed to create spire entry")
		return "", err
	}
	reqLogger.Info("Created entry", "entryID", regEntryId.Id, "spiffeID", spiffeId)

	return regEntryId.Id, nil
}
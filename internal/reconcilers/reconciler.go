package reconcilers

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
)

type Reconciler interface {
	Configuration() *protoapi.NewReconciler
	Name() string
	Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error
	Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error
}

type ReconcileRequest struct {
	CorrelationID string
	TraceID       string
	TeamSlug      string
}

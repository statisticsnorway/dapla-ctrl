package reconcilers

import (
	"context"

	"github.com/statisticsnorway/dapla-api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-api/pkg/apiclient/protoapi"
	"github.com/sirupsen/logrus"
)

type Reconciler interface {
	Configuration() *protoapi.NewReconciler
	Name() string
	Reconcile(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error
	Delete(ctx context.Context, client *apiclient.APIClient, naisTeam *protoapi.Team, log logrus.FieldLogger) error
}

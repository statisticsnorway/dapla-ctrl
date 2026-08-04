package groupserviceaccounts

import (
	"context"

	"google.golang.org/api/iam/v1"
)

type GroupServiceAccounts interface {
	GetOrCreate(ctx context.Context, name, description, projectId string) (*iam.ServiceAccount, error)
	UpdateDescription(ctx context.Context, saName, description string) error
}

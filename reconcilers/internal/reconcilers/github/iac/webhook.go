package iac

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/go-github/v92/github"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

type webhookReconcilerClient interface {
	ListHooks(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.Hook, *github.Response, error)
	CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error)
	EditHook(ctx context.Context, owner, repo string, id int64, body *github.Hook) (*github.Hook, *github.Response, error)
	PingHook(ctx context.Context, owner, repo string, id int64) (*github.Response, error)
	ListHookDeliveries(ctx context.Context, owner, repo string, id int64, opts *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error)
	GetHookDelivery(ctx context.Context, owner, repo string, hookID, deliveryID int64) (*github.HookDelivery, *github.Response, error)
}

type webhookReconciler struct {
	client webhookReconcilerClient
}

// create or update webhook towards atlantis
func (r *webhookReconciler) upsertAtlantisWebhook(ctx context.Context, owner, repo, atlantisUrl, secret string) error {
	hook, err := r.findWebhook(ctx, owner, repo, atlantisUrl)
	if err != nil {
		return err
	}
	if hook == nil {
		return r.createWebHook(ctx, owner, repo, atlantisUrl, secret)
	}
	return r.updateGhRepoAtlantisWebhookSecret(ctx, owner, repo, *hook, secret)
}

// will return nil, nil if the webhook does not exists
func (r *webhookReconciler) findWebhook(ctx context.Context, owner, repo, atlantisUrl string) (*github.Hook, error) {
	hooks, _, err := r.client.ListHooks(ctx, owner, repo, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}

	var hook *github.Hook
	for _, h := range hooks {
		if h.GetConfig().GetURL() == atlantisUrl {
			hook = h
			break
		}
	}
	if hook == nil {
		return nil, nil
	}

	return hook, nil
}

func (r *webhookReconciler) createWebHook(ctx context.Context, owner, repo, atlantisUrl, secret string) error {
	_, _, err := r.client.CreateHook(ctx, owner, repo, &github.Hook{
		Config: &github.HookConfig{
			ContentType: new("json"),
			URL:         &atlantisUrl,
			Secret:      &secret,
		},
		Events: []string{
			"check_run",
			"create",
			"delete",
			"issue_comment",
			"issues",
			"pull_request",
			"pull_request_review",
			"pull_request_review_comment",
			"push",
		},
		Active: new(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create webhook for repo %s: %w", repo, err)
	}
	return nil
}

func (r *webhookReconciler) updateGhRepoAtlantisWebhookSecret(ctx context.Context, owner, repoName string, hook github.Hook, secret string) error {
	hookId := hook.GetID()

	// compare if we need to update secret
	upToDate, err := r.hookLatestDeliverySignedWith(ctx, owner, repoName, hookId, secret)
	if err != nil {
		return err
	}
	if upToDate {
		return nil
	}

	config := *hook.GetConfig()
	config.Secret = &secret
	if _, _, err := r.client.EditHook(ctx, owner, repoName, hookId, &github.Hook{Config: &config}); err != nil {
		return fmt.Errorf("update atlantis webhook secret for repo %s: %w", repoName, err)
	}

	// Trigger webhook so the next reconcile can verify it without updating again.
	if _, err := r.client.PingHook(ctx, owner, repoName, hookId); err != nil {
		return fmt.Errorf("ping atlantis webhook for repo %s: %w", repoName, err)
	}

	return nil
}

// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
// return false if the secret can not be verifed (can happn for no triggers), which implies that the secret should be updated
func (r *webhookReconciler) hookLatestDeliverySignedWith(ctx context.Context, owner, repoName string, hookID int64, secret string) (bool, error) {
	deliveries, _, err := r.client.ListHookDeliveries(ctx, owner, repoName, hookID, &github.ListCursorOptions{First: 1})
	if err != nil {
		return false, err
	}
	if len(deliveries) == 0 {
		return false, nil
	}

	delivery, _, err := r.client.GetHookDelivery(ctx, owner, repoName, hookID, deliveries[0].GetID())
	if err != nil {
		return false, err
	}
	if delivery.Request == nil || delivery.Request.RawPayload == nil {
		return false, nil
	}

	ghSignature := delivery.Request.GetHeader("X-Hub-Signature-256")
	got, err := hex.DecodeString(strings.TrimPrefix(ghSignature, "sha256="))
	if ghSignature == "" || err != nil {
		return false, nil
	}

	signature := generateSignature([]byte(secret), []byte(*delivery.Request.RawPayload))
	return hmac.Equal(signature, got), nil
}

func generateSignature(secret, payload []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return mac.Sum(nil)
}

func getAtlantisUrl(ctx context.Context, client protoapi.AtlantisClient, teamSlug string) (string, error) {
	resp, err := client.GetTeamAtlantis(ctx, &protoapi.GetTeamAtlantisRequest{
		TeamSlug: teamSlug,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get team atlantis config: %w", err)
	}
	customAtlantisName := resp.Config.GetCustomName()

	atlantisName := "atlantis-" + teamSlug
	if customAtlantisName != "" {
		atlantisName = customAtlantisName
	}

	return "https://" + atlantisName + ".atlantis.ssb.no/events", nil
}

func getAtlantisWebhookSecret(ctx context.Context, client protoapi.AtlantisClient, teamSlug string) (string, error) {
	resp, err := client.GetTeamAtlantisWebhookSecret(ctx, &protoapi.GetTeamAtlantisWebhookSecretRequest{
		TeamSlug: teamSlug,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch atlantis webhook secret for team %s: %w", teamSlug, err)
	}
	secret := resp.WebhookSecret
	if secret == "" {
		return "", fmt.Errorf("atlantis webhook secret for team %s is empty", teamSlug)
	}
	return secret, nil
}

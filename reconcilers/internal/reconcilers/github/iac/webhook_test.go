package iac

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/go-github/v92/github"
)

type fakeWebhookClient struct {
	hooks    []*github.Hook
	delivery *github.HookDelivery

	created  *github.Hook
	editedID int64
	edited   *github.Hook
	pingedID int64
}

func (f *fakeWebhookClient) ListHooks(context.Context, string, string, *github.ListOptions) ([]*github.Hook, *github.Response, error) {
	return f.hooks, nil, nil
}

func (f *fakeWebhookClient) CreateHook(_ context.Context, _, _ string, hook *github.Hook) (*github.Hook, *github.Response, error) {
	f.created = hook
	return hook, nil, nil
}

func (f *fakeWebhookClient) EditHook(_ context.Context, _, _ string, id int64, body *github.Hook) (*github.Hook, *github.Response, error) {
	f.editedID = id
	f.edited = body
	return body, nil, nil
}

func (f *fakeWebhookClient) PingHook(_ context.Context, _, _ string, id int64) (*github.Response, error) {
	f.pingedID = id
	return nil, nil
}

func (f *fakeWebhookClient) ListHookDeliveries(context.Context, string, string, int64, *github.ListCursorOptions) ([]*github.HookDelivery, *github.Response, error) {
	if f.delivery == nil {
		return nil, nil, nil
	}
	return []*github.HookDelivery{f.delivery}, nil, nil
}

func (f *fakeWebhookClient) GetHookDelivery(_ context.Context, _, _ string, _, deliveryID int64) (*github.HookDelivery, *github.Response, error) {
	if f.delivery == nil || f.delivery.GetID() != deliveryID {
		return nil, nil, errors.New("delivery not found")
	}
	return f.delivery, nil, nil
}

const (
	testAtlantisURL = "https://atlantis.example.com/events"
	testSecret      = "top-secret"
	testHookID      = int64(42)
)

func atlantisHook() *github.Hook {
	return &github.Hook{
		ID: new(testHookID),
		Config: &github.HookConfig{
			URL:         new(testAtlantisURL),
			ContentType: new("json"),
		},
	}
}

func signedDelivery(secret, signatureHeader string) *github.HookDelivery {
	payload := json.RawMessage(`{"payload":"Hello, World!"}`)
	signature := "sha256=" + hex.EncodeToString(generateSignature([]byte(secret), payload))
	return &github.HookDelivery{
		ID: new(int64(7)),
		Request: &github.HookRequest{
			Headers:    map[string]string{signatureHeader: signature},
			RawPayload: &payload,
		},
	}
}

func TestGenerateSignature(t *testing.T) {
	// https://docs.github.com/en/enterprise-cloud@latest/webhooks/using-webhooks/validating-webhook-deliveries
	expected := "757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
	resultByte := generateSignature([]byte("It's a Secret to Everybody"), []byte("Hello, World!"))
	result := hex.EncodeToString(resultByte)

	if result != expected {
		t.Errorf("expected %s, but got %s", expected, result)
	}
}

func TestUpsertAtlantisWebhook(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		hooks           []*github.Hook
		delivery        *github.HookDelivery
		wantCreate      bool
		wantEdit        bool
		wantTriggerHook bool
	}{
		{
			name:       "creates hook when repo has none",
			wantCreate: true,
		},
		{
			name: "creates hook when only unrelated hooks exist",
			hooks: []*github.Hook{{
				ID:     new(int64(1)),
				Config: &github.HookConfig{URL: new("https://other-hook.example.com")},
			}},
			wantCreate: true,
		},
		{
			name:     "do not update hook when latest delivery is signed with current secret",
			hooks:    []*github.Hook{atlantisHook()},
			delivery: signedDelivery(testSecret, "X-Hub-Signature-256"),
		},
		{
			name:     "case-insensitive signature header lookup ",
			hooks:    []*github.Hook{atlantisHook()},
			delivery: signedDelivery(testSecret, "x-hub-signature-256"),
		},
		{
			name:            "should update webhook when secret is new",
			hooks:           []*github.Hook{atlantisHook()},
			delivery:        signedDelivery("old-secret", "X-Hub-Signature-256"),
			wantEdit:        true,
			wantTriggerHook: true,
		},
		{
			name:            "updates webhook when no deliveries",
			hooks:           []*github.Hook{atlantisHook()},
			wantEdit:        true,
			wantTriggerHook: true,
		},
		{
			name:            "updates webhook when delivery has no signature header",
			hooks:           []*github.Hook{atlantisHook()},
			delivery:        signedDelivery(testSecret, "X-Other-Header"),
			wantEdit:        true,
			wantTriggerHook: true,
		},
		{
			name:  "updates webhook when delivery has no payload",
			hooks: []*github.Hook{atlantisHook()},
			delivery: &github.HookDelivery{
				ID:      new(int64(7)),
				Request: &github.HookRequest{Headers: map[string]string{}},
			},
			wantEdit:        true,
			wantTriggerHook: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &fakeWebhookClient{
				hooks:    tt.hooks,
				delivery: tt.delivery,
			}
			r := &webhookReconciler{client: client}

			err := r.upsertAtlantisWebhook(t.Context(), "statisticsnorway", "hook-iac", testAtlantisURL, testSecret)
			if err != nil {
				t.Fatalf("err = %v", err)
			}

			if got := client.created != nil; got != tt.wantCreate {
				t.Errorf("created = %v, want %v", got, tt.wantCreate)
			}
			if got := client.edited != nil; got != tt.wantEdit {
				t.Errorf("edited = %v, want %v", got, tt.wantEdit)
			}
			if got := client.pingedID != 0; got != tt.wantTriggerHook {
				t.Errorf("pinged = %v, want %v", got, tt.wantTriggerHook)
			}

			if tt.wantCreate {
				hook := client.created
				cfg := hook.GetConfig()
				if cfg.GetURL() != testAtlantisURL || cfg.GetSecret() != testSecret || cfg.GetContentType() != "json" {
					t.Errorf("unexpected hook config: url=%q secret=%q contentType=%q", cfg.GetURL(), cfg.GetSecret(), cfg.GetContentType())
				}
				if !hook.GetActive() {
					t.Error("hook should be active")
				}
				for _, event := range []string{"pull_request", "issue_comment", "push"} {
					if !slices.Contains(hook.Events, event) {
						t.Errorf("event %s was was not found in hooks config", event)
					}
				}
			}

			if tt.wantEdit {
				if client.editedID != testHookID {
					t.Errorf("edited hook id = %d, want %d", client.editedID, testHookID)
				}
				cfg := client.edited.GetConfig()
				if cfg.GetSecret() != testSecret {
					t.Errorf("edited secret = %q, want %q", cfg.GetSecret(), testSecret)
				}
				if cfg.GetURL() != testAtlantisURL || cfg.GetContentType() != "json" {
					t.Errorf("edit should keep existing config, got url=%q contentType=%q", cfg.GetURL(), cfg.GetContentType())
				}
			}

			if tt.wantTriggerHook && client.pingedID != testHookID {
				t.Errorf("pinged hook id = %d, want %d", client.pingedID, testHookID)
			}
		})
	}
}

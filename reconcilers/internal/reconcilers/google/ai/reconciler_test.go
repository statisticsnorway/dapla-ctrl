package ai

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"cloud.google.com/go/billing/budgets/apiv1/budgetspb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/google/go-cmp/cmp"
)

// Reconciling Vertex AI IAM bindings:
// - Preserves an existing unrelated member and IAM binding.
// - A second run does not write the policy again (idempotent)
// - When disabling only the team members are removed.
func TestReconcileAIPlatformUserBindingPreservesUnrelatedAccessAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	r := &reconciler{
		AIPlatformUserRole:  "roles/aiplatform.user",
		GroupSANameTemplate: "developers@example.iam.gserviceaccount.com",
	}
	server := &fakeGoogleServer{policy: &iampb.Policy{Bindings: []*iampb.Binding{
		{Role: r.AIPlatformUserRole, Members: []string{"user:other@example.com"}},
		{Role: "roles/viewer", Members: []string{"group:auditors@example.com"}},
	}}}
	projects, _ := fakeGoogleClients(t, server)

	if err := reconcileAIPlatformUserBinding(r, ctx, projects, "my-team", "project-id", true); err != nil {
		t.Fatal(err)
	}
	wantMembers := []string{
		"user:other@example.com",
		"group:my-team-developers@groups.ssb.no",
		"serviceAccount:my-team-developers@example.iam.gserviceaccount.com",
	}
	if diff := cmp.Diff(wantMembers, server.policy.Bindings[0].Members); diff != "" {
		t.Fatalf("AI binding members mismatch (-want +got):\n%s", diff)
	}
	if server.policy.Bindings[1].Role != "roles/viewer" {
		t.Fatal("unrelated binding was changed")
	}

	if err := reconcileAIPlatformUserBinding(r, ctx, projects, "my-team", "project-id", true); err != nil {
		t.Fatal(err)
	}
	if server.setPolicyCalls != 1 {
		t.Fatalf("idempotent reconciliation set policy %d times, want 1", server.setPolicyCalls)
	}

	if err := reconcileAIPlatformUserBinding(r, ctx, projects, "my-team", "project-id", false); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff([]string{"user:other@example.com"}, server.policy.Bindings[0].Members); diff != "" {
		t.Fatalf("disable removed unrelated member (-want +got):\n%s", diff)
	}
}

// Reconciling budget notification channels:
// - Creates the appropriate amount of channels
// - A second run doesn't re-create the channels (idempotency)
// - Disabling channels deletes the appropriate channels
func TestReconcileAIBudgetNotificationChannelsConvergesAndCleansUp(t *testing.T) {
	ctx := context.Background()
	r := &reconciler{
		AIBudgetNotificationName: "Vertex AI budget notification channel",
	}
	server := &fakeGoogleServer{channels: []*monitoringpb.NotificationChannel{
		{Name: "projects/project-id/notificationChannels/existing", DisplayName: r.AIBudgetNotificationName, Type: aiBudgetNotificationType, Labels: map[string]string{aiBudgetNotificationLabel: "one@example.com"}},
		{Name: "projects/project-id/notificationChannels/unrelated", DisplayName: "Unrelated", Type: aiBudgetNotificationType, Labels: map[string]string{aiBudgetNotificationLabel: "other@example.com"}},
	}}
	_, channels := fakeGoogleClients(t, server)
	emails := []string{"one@example.com", "two@example.com"}

	names, err := reconcileAIBudgetNotificationChannels(r, ctx, channels, "project-id", emails)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != len(emails) || server.createChannelCalls != 1 {
		t.Fatalf("got %d channel names and %d creates, want 2 names and 1 create", len(names), server.createChannelCalls)
	}

	if _, err := reconcileAIBudgetNotificationChannels(r, ctx, channels, "project-id", emails); err != nil {
		t.Fatal(err)
	}
	if server.createChannelCalls != 1 {
		t.Fatalf("idempotent reconciliation created %d channels, want 1", server.createChannelCalls)
	}

	if _, err := reconcileAIBudgetNotificationChannels(r, ctx, channels, "project-id", nil); err != nil {
		t.Fatal(err)
	}
	if server.deleteChannelCalls != 2 {
		t.Fatalf("deleted %d AI channels, want 2", server.deleteChannelCalls)
	}
	if len(server.channels) != 1 || server.channels[0].DisplayName != "Unrelated" {
		t.Fatal("cleanup removed an unrelated notification channel")
	}
}

func TestGetAIBudget(t *testing.T) {
	r := &reconciler{
		AIBudgetThresholds: []float64{0.5, 0.9, 1.0},
	}
	budgetLimit := int64(100)
	notificationChannels := []string{"channels/one"}
	projectNumber := "12345"
	budget := getAIBudget(r, "my-team", projectNumber, budgetLimit, notificationChannels)

	if budget.DisplayName != "my-team AI budget" ||
		!slices.Equal(budget.BudgetFilter.Projects, []string{fmt.Sprintf("projects/%s", projectNumber)}) ||
		!slices.Equal(budget.BudgetFilter.Services, []string{vertexAIServiceName}) ||
		budget.BudgetFilter.GetCalendarPeriod() != budgetspb.CalendarPeriod_MONTH ||
		budget.Amount.GetSpecifiedAmount().CurrencyCode != aiBudgetCurrencyCode ||
		budget.Amount.GetSpecifiedAmount().Units != budgetLimit ||
		!slices.EqualFunc(budget.ThresholdRules, r.AIBudgetThresholds, func(rule *budgetspb.ThresholdRule, threshold float64) bool { return rule.ThresholdPercent == threshold }) ||
		!budget.NotificationsRule.DisableDefaultIamRecipients ||
		!slices.Equal(budget.NotificationsRule.MonitoringNotificationChannels, notificationChannels) {
		t.Fatalf("unexpected budget: %v", budget)
	}
}

func TestGetStandardProjectID(t *testing.T) {
	server := &fakeGoogleServer{projects: []*resourcemanagerpb.Project{
		{ProjectId: "unrelated-t-123"},
		{ProjectId: "my-team-troika-t-231"},
		{ProjectId: "my-teaser-team-t-398"},
		{ProjectId: "my-team-t-456"},
		{ProjectId: "my-team-trace-t-fu"},
	}}
	projects, _ := fakeGoogleClients(t, server)

	got, err := getStandardProjectID(context.Background(), projects, "folder-id", "my-team", "test")
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-team-t-456" {
		t.Fatalf("got project %q, want %q", got, "my-team-t-456")
	}
}

package googleresourcemanager

import (
	"reflect"
	"testing"

	"google.golang.org/api/cloudresourcemanager/v1"
)

func TestAddBindingsToPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      *cloudresourcemanager.Policy
		member      string
		roles       []string
		wantChanged bool
		wantPolicy  map[string][]string
	}{
		{
			name: "adds member to existing binding",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com", "user:alice@example.com"},
			},
		},
		{
			name: "adds new binding for missing role",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/editor"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com"},
				"roles/editor": {"user:alice@example.com"},
			},
		},
		{
			name: "existing member is no-op",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com", "user:bob@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: false,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:alice@example.com", "user:bob@example.com"},
			},
		},
		{
			name:        "duplicate role inputs add member once",
			policy:      &cloudresourcemanager.Policy{},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer", "roles/viewer"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:alice@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChanged := addBindingsToPolicy(tt.policy, tt.member, tt.roles...)
			if gotChanged != tt.wantChanged {
				t.Fatalf("addBindingsToPolicy() changed = %t, want %t", gotChanged, tt.wantChanged)
			}

			if gotPolicy := bindingsByRole(tt.policy); !reflect.DeepEqual(gotPolicy, tt.wantPolicy) {
				t.Fatalf("policy bindings = %#v, want %#v", gotPolicy, tt.wantPolicy)
			}
		})
	}
}

func TestAddBindingsToPolicyIsIdempotent(t *testing.T) {
	policy := &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
		{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
	}}

	if changed := addBindingsToPolicy(policy, "user:alice@example.com", "roles/viewer", "roles/editor"); !changed {
		t.Fatal("addBindingsToPolicy() changed = false, want true")
	}
	expectedPolicyAfterFirstAdd := bindingsByRole(policy)

	if changed := addBindingsToPolicy(policy, "user:alice@example.com", "roles/viewer", "roles/editor"); changed {
		t.Fatal("addBindingsToPolicy() changed = true on second add, want false")
	}
	if gotPolicy := bindingsByRole(policy); !reflect.DeepEqual(gotPolicy, expectedPolicyAfterFirstAdd) {
		t.Fatalf("second add changed policy bindings = %#v, want %#v", gotPolicy, expectedPolicyAfterFirstAdd)
	}
}

func TestRemoveMemberFromPolicy(t *testing.T) {
	tests := []struct {
		name        string
		policy      *cloudresourcemanager.Policy
		member      string
		roles       []string
		wantChanged bool
		wantPolicy  map[string][]string
	}{
		{
			name: "removes member from binding",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com", "user:bob@example.com"}},
				{Role: "roles/editor", Members: []string{"user:carol@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com"},
				"roles/editor": {"user:carol@example.com"},
			},
		},
		{
			name: "removes binding when member is last member",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com"}},
				{Role: "roles/editor", Members: []string{"user:carol@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/editor": {"user:carol@example.com"},
			},
		},
		{
			name: "missing role is no-op",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/editor"},
			wantChanged: false,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com"},
			},
		},
		{
			name: "missing member is no-op",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: false,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com"},
			},
		},
		{
			name: "removes duplicate memberships",
			policy: &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com", "user:bob@example.com", "user:alice@example.com"}},
			}},
			member:      "user:alice@example.com",
			roles:       []string{"roles/viewer"},
			wantChanged: true,
			wantPolicy: map[string][]string{
				"roles/viewer": {"user:bob@example.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChanged := removeMemberFromPolicy(tt.policy, tt.member, tt.roles...)
			if gotChanged != tt.wantChanged {
				t.Fatalf("removeMemberFromPolicy() changed = %t, want %t", gotChanged, tt.wantChanged)
			}

			if gotPolicy := bindingsByRole(tt.policy); !reflect.DeepEqual(gotPolicy, tt.wantPolicy) {
				t.Fatalf("policy bindings = %#v, want %#v", gotPolicy, tt.wantPolicy)
			}
		})
	}
}

func TestRemoveMemberFromPolicyIsIdempotent(t *testing.T) {
	policy := &cloudresourcemanager.Policy{Bindings: []*cloudresourcemanager.Binding{
		{Role: "roles/viewer", Members: []string{"user:alice@example.com", "user:bob@example.com"}},
	}}

	if changed := removeMemberFromPolicy(policy, "user:alice@example.com", "roles/viewer"); !changed {
		t.Fatal("removeMemberFromPolicy() changed = false, want true")
	}
	expectedPolicyAfterFirstRemoval := bindingsByRole(policy)

	if changed := removeMemberFromPolicy(policy, "user:alice@example.com", "roles/viewer"); changed {
		t.Fatal("removeMemberFromPolicy() changed = true on second removal, want false")
	}
	if gotPolicy := bindingsByRole(policy); !reflect.DeepEqual(gotPolicy, expectedPolicyAfterFirstRemoval) {
		t.Fatalf("second removal changed policy bindings = %#v, want %#v", gotPolicy, expectedPolicyAfterFirstRemoval)
	}
}

func bindingsByRole(policy *cloudresourcemanager.Policy) map[string][]string {
	bindings := make(map[string][]string, len(policy.Bindings))
	for _, binding := range policy.Bindings {
		bindings[binding.Role] = append([]string(nil), binding.Members...)
	}
	return bindings
}

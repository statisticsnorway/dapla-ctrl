package parquedit

import (
	"testing"

	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

func TestEnvironmentRegistrationMetadata(t *testing.T) {
	prod, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	test, err := New(t.Context(), WithEnvironment(Test))
	if err != nil {
		t.Fatal(err)
	}

	prodConfig := prod.Configuration()
	testConfig := test.Configuration()
	if prodConfig.Name != "dapla:ffunk:parquedit" || prodConfig.DisplayName != "Parquedit" {
		t.Fatalf("production identity changed: name=%q displayName=%q", prodConfig.Name, prodConfig.DisplayName)
	}
	if testConfig.Name != "dapla:ffunk:parquedit:test" || testConfig.DisplayName != "Parquedit (test)" {
		t.Fatalf("unexpected test identity: name=%q displayName=%q", testConfig.Name, testConfig.DisplayName)
	}
	if prodConfig.DisplayName == testConfig.DisplayName {
		t.Fatal("reconciler display names must be distinct for database registration")
	}
	for _, cfg := range []*reconciler{prod, test} {
		spec := cfg.Configuration()
		if spec.Name != cfg.Name() {
			t.Fatalf("configuration name %q differs from reconciler name %q", spec.Name, cfg.Name())
		}
		if spec.EnableByDefault {
			t.Fatalf("%s must not enable itself by default", spec.Name)
		}
		if len(spec.Config) != 2 ||
			spec.Config[0].Key != "labReconcilerEndpoint" ||
			spec.Config[0].Secret ||
			spec.Config[1].Key != "labReconcilerSecret" ||
			!spec.Config[1].Secret {
			t.Fatalf("configuration keys or secret flags changed for %s", spec.Name)
		}
	}
}

func TestDesiredStateUsesIndependentTeamFlags(t *testing.T) {
	for _, tt := range []struct {
		name        string
		env         Environment
		prodEnabled bool
		testEnabled bool
		want        bool
	}{
		{name: "both disabled in prod", env: Prod},
		{name: "prod only", env: Prod, prodEnabled: true, want: true},
		{name: "prod disabled while test enabled", env: Prod, testEnabled: true},
		{name: "both enabled in prod", env: Prod, prodEnabled: true, testEnabled: true, want: true},
		{name: "both disabled in test", env: Test},
		{name: "test only", env: Test, testEnabled: true, want: true},
		{name: "test disabled while prod enabled", env: Test, prodEnabled: true},
		{name: "both enabled in test", env: Test, prodEnabled: true, testEnabled: true, want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(t.Context(), WithEnvironment(tt.env))
			if err != nil {
				t.Fatal(err)
			}
			got := r.desiredState(&protoapi.Team{
				HasParquedit:     tt.prodEnabled,
				HasParqueditTest: tt.testEnabled,
			})
			if got != tt.want {
				t.Fatalf("desiredState() = %v, want %v", got, tt.want)
			}
		})
	}
}

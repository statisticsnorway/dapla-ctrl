package pam_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	reconcilerconfig "github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/cmd/reconciler/config"
	googlepam "github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/google/pam"
)

func examplePAMCondition(t *testing.T) string {
	t.Helper()
	example, err := os.ReadFile("../../../../.env.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(example), "\n") {
		if !strings.HasPrefix(line, "# PAM_MANAGED_TEAM_FOLDER_CONDITION_TEMPLATE=") {
			continue
		}
		values, err := godotenv.Unmarshal(strings.TrimPrefix(line, "# "))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := reconcilerconfig.NewConfig(t.Context(), envconfig.MapLookuper(values))
		if err != nil {
			t.Fatal(err)
		}
		return cfg.PAMManagedTeamFolder.ConditionTemplate
	}
	t.Fatal("condition template example is missing")
	return ""
}

func conditionProgram(t *testing.T, expression string) cel.Program {
	t.Helper()
	env, err := cel.NewEnv(
		cel.Variable("resource", cel.MapType(cel.StringType, cel.StringType)),
		cel.Function("extract", cel.MemberOverload("iam_bucket_extract",
			[]*cel.Type{cel.StringType, cel.StringType}, cel.StringType,
			cel.BinaryBinding(func(value, extractionTemplate ref.Val) ref.Val {
				// IAM adds extract to CEL. Model the two bucket templates used by the example.
				name, ok := strings.CutPrefix(string(value.(types.String)), "projects/_/buckets/")
				if !ok {
					return types.String("")
				}
				switch string(extractionTemplate.(types.String)) {
				case "projects/_/buckets/{bucket}":
					return types.String(name)
				case "projects/_/buckets/{bucket}/":
					bucket, _, found := strings.Cut(name, "/")
					if !found {
						return types.String("")
					}
					return types.String(bucket)
				default:
					return types.NewErr("unsupported extraction template")
				}
			}))),
	)
	if err != nil {
		t.Fatal(err)
	}
	ast, issues := env.Compile(expression)
	if issues.Err() != nil {
		t.Fatalf("invalid CEL expression: %v\n%s", issues.Err(), expression)
	}
	program, err := env.Program(ast)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestPAMConditionExample_ScopesBucketAndObjectAccess(t *testing.T) {
	server := newFakeGcpTeamResourcesServer()
	pam := &fakePrivilegedAccessManager{}
	r, err := googlepam.New(t.Context(), googlepam.Config{
		Environments:      []string{"dev", "test", "prod"},
		Role:              "roles/storage.objectViewer",
		Principals:        []string{"user:reader@example.com"},
		MaxGrantDuration:  time.Hour,
		ConditionTemplate: examplePAMCondition(t),
	}, nil, googlepam.WithPrivilegedAccessManager(pam))
	if err != nil {
		t.Fatal(err)
	}
	apiClient := startFakeGrpcServer(t, server)
	for _, team := range []string{"my-team", "another-team"} {
		for _, env := range []string{"dev", "test", "prod"} {
			server.folders[team+"/"+env] = &protoapi.GcpTeamFolder{TeamSlug: team, Env: env, FolderId: team + "/" + env}
		}
		if err := r.Reconcile(t.Context(), apiClient, &protoapi.Team{Slug: team, IsManaged: true}, logrus.New()); err != nil {
			t.Fatal(err)
		}
	}
	if len(pam.calls) != 6 {
		t.Fatalf("got %d entitlements, want 6", len(pam.calls))
	}
	for _, call := range pam.calls {
		team, env, _ := strings.Cut(call.folderID, "/")
		t.Run(team+"/"+env, func(t *testing.T) {
			program := conditionProgram(t, call.conditionExpression)
			prefix := "ssb-" + team + "-data-"
			otherEnv := "prod"
			if env == "prod" {
				otherEnv = "dev"
			}
			tests := map[string]struct {
				bucket string
				want   bool
			}{
				"produkt":                     {bucket: prefix + "produkt-" + env, want: true},
				"delt":                        {bucket: prefix + "delt-stat-" + env, want: true},
				"delt hyphenated name":        {bucket: prefix + "delt-kort-navn-" + env, want: true},
				"delomat":                     {bucket: prefix + "delt-delomat-stat-" + env, want: true},
				"kilde":                       {bucket: prefix + "kilde-" + env},
				"another team":                {bucket: "ssb-unrelated-data-produkt-" + env},
				"similar team":                {bucket: "ssb-" + team + "-extra-data-delt-stat-" + env},
				"another environment":         {bucket: prefix + "produkt-" + otherEnv},
				"delt another environment":    {bucket: prefix + "delt-stat-" + otherEnv},
				"delomat another environment": {bucket: prefix + "delt-delomat-stat-" + otherEnv},
				"produkt suffix":              {bucket: prefix + "produkt-extra-" + env},
				"environment suffix":          {bucket: prefix + "produkt-" + env + "-extra"},
				"missing short name":          {bucket: prefix + "delt--" + env},
				"missing delomat short name":  {bucket: prefix + "delt-delomat--" + env},
				"missing environment":         {bucket: prefix + "delt-stat"},
				"other bucket family":         {bucket: prefix + "annet-" + env},
			}
			for name, tt := range tests {
				t.Run(name, func(t *testing.T) {
					for _, resourceType := range []string{"storage.googleapis.com/Bucket", "storage.googleapis.com/Object"} {
						resourceName := "projects/_/buckets/" + tt.bucket
						if resourceType == "storage.googleapis.com/Object" {
							resourceName += "/objects/nested/object.json"
						}
						assertConditionResult(t, program, resourceType, resourceName, tt.want)
					}
				})
			}
			t.Run("object name cannot spoof bucket", func(t *testing.T) {
				resourceName := fmt.Sprintf("projects/_/buckets/%skilde-%s/objects/projects/_/buckets/%sprodukt-%s/objects/example", prefix, env, prefix, env)
				assertConditionResult(t, program, "storage.googleapis.com/Object", resourceName, false)
				resourceName = fmt.Sprintf("projects/_/buckets/%sdelt-stat-%s/objects/spoof-%s", prefix, otherEnv, env)
				assertConditionResult(t, program, "storage.googleapis.com/Object", resourceName, false)
			})
			t.Run("unrelated resource type", func(t *testing.T) {
				assertConditionResult(t, program, "compute.googleapis.com/Disk", "projects/_/buckets/"+prefix+"produkt-"+env, false)
			})
		})
	}
}

func assertConditionResult(t *testing.T, program cel.Program, resourceType, resourceName string, want bool) {
	t.Helper()
	result, _, err := program.Eval(map[string]any{
		"resource": map[string]string{"type": resourceType, "name": resourceName},
	})
	if err != nil {
		t.Fatalf("evaluate condition for %s: %v", resourceName, err)
	}
	if result != types.Bool(want) {
		t.Errorf("%s %s: got %v, want %t", resourceType, resourceName, result, want)
	}
}

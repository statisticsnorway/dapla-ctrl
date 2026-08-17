package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	goose.AddMigrationContext(Up0015, Down0015)
}

func Up0015(ctx context.Context, tx *sql.Tx) error {
	for team, repos := range getTeamsAndArtifactRegistryGithubRepos() {
		exists, err := teamExists(tx, team)
		if err != nil {
			return err
		}
		if !exists {
			log.WithField("team", team).Warnf("not found in db. skipping insert into team_artifact_registry_repositories")
			continue
		}

		_, err = tx.ExecContext(ctx, fmt.Sprintf(`INSERT INTO team_artifact_registry_repositories(team_slug, format, size_bytes) VALUES ('%s', 'docker', 0);`, team))
		if err != nil {
			return err
		}

		ghReposValues := make([]string, len(repos))
		for i, repo := range repos {
			ghReposValues[i] = fmt.Sprintf(`('%s', '%s')`, team, repo)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO team_artifact_registry_gh_repos_allow_list (team_slug, repository_name)
			VALUES
				`+strings.Join(ghReposValues, ",")+`
		 ;`)
		if err != nil {
			return err
		}
	}

	return nil
}

func teamExists(tx *sql.Tx, teamSlug string) (bool, error) {
	var exists bool
	err := tx.QueryRow("SELECT EXISTS (SELECT 1 FROM teams WHERE slug = $1)", teamSlug).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	return exists, nil
}

func Down0015(ctx context.Context, tx *sql.Tx) error {
	return nil
}

// Fetched for terraform-ssb-dapla-teams per 13.08.2026
// with command:
// find teams -type f -name 'team.yaml' | xargs -I {} yq '.uniform_name, .artifact_registry.repos' "{}"
func getTeamsAndArtifactRegistryGithubRepos() map[string][]string {
	return map[string][]string{
		"a200-stoetteteam": {
			"a200-stoetteteam-iac",
		},
		"a300-gis-pii": {
			"a300-gis-pii-iac",
		},
		"aordning-register": {
			"aordning-register-iac",
		},
		"arbmark-aku": {
			"arbmark-aku-iac",
		},
		"arbmark-register": {
			"arbmark-register-iac",
		},
		"arbmark-skjema": {
			"arbmark-skjema-iac",
		},
		"arbulykker": {
			"arbulykker-iac",
		},
		"areal": {
			"areal-iac",
		},
		"arkitektur": {
			"arkitektur-iac",
		},
		"barnevern": {
			"barn-bvr-api",
			"barn-bufdir-test",
			"barn-report-gateway",
			"barn-eventstore",
			"barn-fiksio-transceiver",
			"barn-snapshot-to-kostra",
			"barn-replay-app",
			"barn-kostra-receiver",
			"barn-api-gateway",
			"barn-maven-parent",
			"barn-pubsub-test",
			"barn-data-generator",
			"barn-pubsub-models",
			"barn-bigquery-api",
			"barn-snapshot-models",
			"barn-barnevern-xsd",
			"barn-kotlin-utils",
			"barn-joachim-onboarding",
			"cloud-storage-commons",
			"barnevern-iac",
		},
		"barneverng": {
			"barneverng-iac",
		},
		"barneverni": {
			"barneverni-iac",
		},
		"bef-analyse": {
			"bef-analyse-iac",
		},
		"bef-bosted": {
			"bef-bosted-iac",
		},
		"bef-husholdning": {
			"bef-husholdning-iac",
		},
		"bef-oppdrag": {
			"bef-oppdrag-iac",
		},
		"bef-regional": {
			"bef-regional-iac",
		},
		"bef-statistikk": {
			"bef-statistikk-iac",
		},
		"bef-valg": {
			"bef-valg-iac",
		},
		"befstat": {
			"befstat-iac",
		},
		"betfor": {
			"betfor-iac",
		},
		"bilpark-nybil": {
			"bilpark-nybil-iac",
		},
		"bki": {
			"bki-iac",
		},
		"bof-stedfesting": {
			"bof-stedfesting-iac",
		},
		"boforholdregister": {
			"boforholdregister-iac",
		},
		"bpi": {
			"bpi-iac",
		},
		"brann-kostra": {
			"brann-kostra-iac",
		},
		"bygg": {
			"bygg-iac",
		},
		"dapla-felles": {
			"dapla-automation-processor",
			"dapla-felles-iac",
			"dapla-automation-scaler",
		},
		"dapla-ffunk": {
			"dapla-ffunk-iac",
		},
		"dapla-lab": {
			"dapla-jdemetra-docker",
			"dapla-lab-images",
			"dapla-lab-helm-image",
			"dapla-lab-iac",
			"dapla-vscode-python",
			"datadoc",
			"jupyterhub-project",
			"keycloakerator",
			"noddy",
			"onyxia",
			"onyxia-api",
			"ssbucketeer",
			"sirius-editering",
			"statbank-pxwebapi",
			"onyxia-janitor",
			"labid",
			"dapla-ctrl",
		},
		"dapla-metadata": {
			"dapla-metadata-iac",
			"datadoc-editor",
			"datadoc-service",
			"vardef",
			"metadata-api-gateway",
			"klass-web",
			"ssb-dataportal",
			"metamapper",
			"dapla-samling-workshop-2025",
		},
		"dapla-onprem-vpn": {
			"dapla-onprem-vpn-iac",
		},
		"dapla-platform": {
			"dapla-gcp-vpc-automations",
			"dapla-team-api",
			"dapla-platform-iac",
			"dapla-kuben-resource-model",
		},
		"dapla-pseudo": {
			"pseudo-service",
			"dapla-pseudo-iac",
			"dapla-dlp-pseudo-service",
			"dapla-dlp-pseudo-func",
			"dapla-dlp-pseudo-core",
			"tink-fpe-java",
		},
		"dapla-skyinfra": {
			"dapla-skyinfra-iac",
			"aig-test",
		},
		"dapla-stat": {
			"pseudo-service",
			"jupyterhub-project",
			"dapla-stat-iac",
			"dapla-ctrl",
			"dapla-statbank-authenticator",
		},
		"datafangst-alf": {
			"datafangst-alf-iac",
		},
		"datafangst-altinn": {
			"datafangst-altinn-iac",
		},
		"datafangst-blaise": {
			"datafangst-blaise-iac",
		},
		"datafangst-innfin": {
			"datafangst-innfin-iac",
		},
		"datafangst-kart": {
			"datafangst-kart-iac",
		},
		"datafangst-m2m": {
			"datafangst-m2m-iac",
		},
		"datafangst-person": {
			"datafangst-person-iac",
		},
		"datafangst-planle": {
			"datafangst-planle-iac",
		},
		"dftjen": {
			"dftjen-iac",
			"maskinporten-guardian",
			"guardian-client-java",
			"kudoc",
			"data-collector-api",
			"data-collector-server-base",
			"data-collector-testutils",
			"data-collector-connector-content-stream-discarding",
			"data-collector-connector-content-stream-rawdata",
			"data-collector-core",
			"data-collector-docker",
			"data-collector-javascript-processor",
			"rawdata-client-provider-gcs",
			"rawdata-client-api",
			"service-provider-api",
			"dapla-secrets-provider-google-secret-manager",
			"data-collector-samples",
			"dapla-secrets-client-api",
			"dapla-secrets-provider-dynamic-configuration",
			"dapla-secrets-provider-safe-configuration",
			"workshop-app-dftjen",
			"data-collector-monitor",
		},
		"eiendoms-kostra": {
			"eiendoms-kostra-iac",
		},
		"el-pris-volum": {
			"el-pris-volum-iac",
		},
		"energ-ind-korttid": {
			"energ-ind-korttid-iac",
		},
		"energi-ereb": {
			"energi-ereb-iac",
		},
		"energi-fjernvarme": {
			"energi-fjernvarme-iac",
		},
		"energi-industri": {
			"energi-industri-iac",
		},
		"energi-kostra": {
			"energi-kostra-iac",
		},
		"energi-petrole": {
			"energi-petrole-iac",
		},
		"energi-rapp": {
			"energi-rapp-iac",
		},
		"etlev": {
			"etlev-iac",
		},
		"famvern": {
			"famvern-iac",
		},
		"fastlegetj": {
			"fastlegetj-iac",
		},
		"finmark": {
			"finmark-iac",
		},
		"finpop": {
			"finpop-iac",
		},
		"finregn": {
			"finregn-iac",
		},
		"folksim": {
			"folksim-iac",
		},
		"forbruk-isolert": {
			"forbruk-isolert-iac",
		},
		"forbruk-stat": {
			"forbruk-stat-iac",
		},
		"forbruk-transak": {
			"forbruk-transak-iac",
		},
		"fordelingsregnska": {
			"fordelingsregnska-iac",
		},
		"forsk-mismatch": {
			"forsk-mismatch-iac",
		},
		"forsk-nhm": {
			"forsk-nhm-iac",
		},
		"forskteknar-forsk": {
			"forskteknar-forsk-iac",
		},
		"forskteknar-narin": {
			"forskteknar-narin-iac",
		},
		"forskteknar-tek": {
			"forskteknar-tek-iac",
		},
		"funkhem": {
			"funkhem-iac",
		},
		"hack-fnrleting": {
			"hack-fnrleting-iac",
		},
		"helseforhold": {
			"helseforhold-iac",
		},
		"helsetjko": {
			"helsetjko-iac",
		},
		"infoforvaltning": {
			"infoforvaltning-iac",
		},
		"inntekt-hushold": {
			"inntekt-hushold-iac",
		},
		"itinfra-bakkesyst": {
			"itinfra-bakkesyst-iac",
			"rstudio-onprem-ghashr",
			"awx_oracle_ee",
			"AWX_Execution_Environment",
			"jupyterhub-onprem-2025",
			"jupyterhub-onprem",
			"jupyterlab-common",
			"awx_ee_builder",
			"awx_ee_builder_oracle",
		},
		"itinfra-ks": {
			"itinfra-ks-iac",
		},
		"keycloak": {
			"keycloak-iac",
		},
		"ki-lab": {
			"ki-lab-iac",
		},
		"kombolig-kostra": {
			"kombolig-kostra-iac",
		},
		"kosthald": {
			"kosthald-iac",
		},
		"kostra": {
			"kostra-iac",
			"kostra-transformer",
			"kostra-kontrollprogram",
			"kostra-commongui",
			"kostra-data",
			"kostra-dockerbuilds",
			"kostra-emailservice",
			"kostra-executionservice",
			"kostra-metadata",
			"kostra-methodlibrary",
			"kostra-oauthserver",
			"kostra-rserve",
			"kostra-system",
			"java-vtl",
			"java-vtl-connectors",
			"java-vtl-tools",
			"ssb-kostra",
		},
		"kostra-pbm": {
			"kostra-pbm-iac",
		},
		"krim-og-rett": {
			"krim-og-rett-iac",
		},
		"kultur-biblmuseum": {
			"kultur-biblmuseum-iac",
		},
		"kultur-kultmed": {
			"kultur-kultmed-iac",
		},
		"kultur-trokirke": {
			"kultur-trokirke-iac",
		},
		"kurs-fra-kilde": {
			"kurs-fra-kilde-iac",
		},
		"laerermod": {
			"laerermod-iac",
		},
		"lb-sky": {
			"lb-sky-iac",
		},
		"levekaar": {
			"levekaar-iac",
		},
		"levekaar-arbeid": {
			"levekaar-arbeid-iac",
		},
		"levekaar-barn": {
			"levekaar-barn-iac",
		},
		"livskvalitet": {
			"livskvalitet-iac",
		},
		"lotte": {
			"lotte-iac",
		},
		"makro-demec": {
			"makro-demec-iac",
		},
		"makro-kvarts": {
			"makro-kvarts-iac",
		},
		"malt": {
			"malt-iac",
		},
		"metode-naceomlegg": {
			"metode-naceomlegg-iac",
		},
		"metode-sdc": {
			"metode-sdc-iac",
		},
		"metode-sesong": {
			"metode-sesong-iac",
		},
		"metode-skjema": {
			"metode-skjema-iac",
		},
		"microdata": {
			"microdata-iac",
			"microdata-data-service",
			"microdata-metadata-service",
			"microdata-pseudonym-service",
			"microdata-job-service",
			"microdata-job-executor",
			"microdata-datastore-admin",
			"microdata-datastore-admin-backend",
			"microdata-datastores-info",
			"microdata-datastore-initializr",
			"microdata-job-db",
			"microdata-nginx",
			"microdata-depseudonymization",
			"microdata-system-test",
			"microdata-alerter",
			"microdata-jenkins",
			"microdata-auditbeat",
			"microdata-elk",
			"microdata-datastore-api",
			"microdata-datastore-encrypter",
			"microdata-system",
		},
		"microdata-no": {
			"microdata-no-iac",
		},
		"mikrodata-oppdrag": {
			"mikrodata-oppdrag-iac",
		},
		"mosart": {
			"mosart-iac",
		},
		"naturregnskap": {
			"naturregnskap-iac",
		},
		"nr": {
			"nr-iac",
		},
		"nr-likningsmodul": {
			"nr-likningsmodul-iac",
		},
		"nr-rekreasjon": {
			"nr-rekreasjon-iac",
		},
		"nspek": {
			"nspek-iac",
		},
		"off-fin": {
			"off-fin-iac",
		},
		"opera-forretning": {
			"opera-forretning-iac",
		},
		"papis": {
			"papis-iac",
		},
		"pasient": {
			"pasient-iac",
		},
		"play-chatbot": {
			"play-chatbot-iac",
			"team-ki-chatbots",
		},
		"play-datafangst": {
			"play-datafangst-iac",
		},
		"play-enhjoern-a": {
			"play-enhjoern-a-iac",
		},
		"play-ffunk-edit-a": {
			"play-ffunk-edit-a-iac",
		},
		"play-ffunk-edit-b": {
			"play-ffunk-edit-b-iac",
		},
		"play-foeniks-a": {
			"play-foeniks-a-iac",
		},
		"play-lab": {
			"play-lab-iac",
		},
		"play-mabl-test": {
			"play-mabl-test-iac",
		},
		"play-obr": {
			"play-obr-iac",
		},
		"play-obr-b": {
			"play-obr-b-iac",
		},
		"play-skatt": {
			"play-skatt-iac",
		},
		"play-skyinfra-a": {
			"play-skyinfra-a-iac",
		},
		"pleie": {
			"pleie-iac",
		},
		"poc-composer": {
			"poc-composer-iac",
		},
		"pop-lb-eiendom": {
			"pop-lb-eiendom-iac",
		},
		"pop-mat-gb": {
			"pop-mat-gb-iac",
		},
		"primaer-fiske": {
			"primaer-fiske-iac",
		},
		"primaer-j-kostra": {
			"primaer-j-kostra-iac",
		},
		"primaer-j-reg": {
			"primaer-j-reg-iac",
		},
		"primaer-j-skjema": {
			"primaer-j-skjema-iac",
		},
		"primaer-jakt": {
			"primaer-jakt-iac",
		},
		"prisstat": {
			"prisstat-iac",
		},
		"proj-aiml-edit": {
			"proj-aiml-edit-iac",
		},
		"pseudo-comp-test": {
			"pseudo-comp-test-iac",
		},
		"reg-a-ordningen": {
			"reg-a-ordningen-iac",
		},
		"reg-bof": {
			"reg-bof-iac",
		},
		"reg-freg": {
			"reg-freg-iac",
		},
		"reg-matrikkel": {
			"reg-matrikkel-iac",
			"reg-grunnbok-harvester",
		},
		"reiseliv-korttid": {
			"reiseliv-korttid-iac",
		},
		"royk": {
			"royk-iac",
		},
		"sikkerhetssenter": {
			"sikkerhetssenter-iac",
		},
		"skatt-naering": {
			"skatt-naering-iac",
		},
		"skatt-person": {
			"skatt-person-iac",
		},
		"sosial-kvp": {
			"sosial-kvp-iac",
		},
		"sosialhjelpk": {
			"sosialhjelpk-iac",
		},
		"speshelse": {
			"speshelse-iac",
		},
		"ssbno": {
			"ssbno-iac",
			"jupyter-ssbno",
		},
		"stat-skog": {
			"stat-skog-iac",
		},
		"statbank": {
			"statbank-iac",
		},
		"statreg-admin": {
			"statreg-admin-iac",
		},
		"strukt-arbkost": {
			"strukt-arbkost-iac",
		},
		"strukt-kraftnarin": {
			"strukt-kraftnarin-iac",
		},
		"strukt-miljoe": {
			"strukt-miljoe-iac",
		},
		"strukt-mva": {
			"strukt-mva-iac",
		},
		"strukt-naering": {
			"strukt-naering-iac",
			"stat-prodcom",
			"stat-naringer-dash",
		},
		"sup": {
			"sup-iac",
		},
		"suv-altinn": {
			"suv-altinn-iac",
			"altinn3-admin-api",
			"altinn3-admin-gui",
			"altinn3-auth-service",
			"altinn3-db-api",
			"altinn3-error-handler",
			"altinn3-event-api",
			"altinn3-event-generator",
			"altinn3-event-processor",
			"altinn3-external-data-api",
			"altinn3-fastapi-template",
			"altinn3-form-monitoring",
			"altinn3-innkvittering-processor",
			"altinn3-instance-processor",
			"altinn3-instantiation-service",
			"altinn3-instantiator",
			"altinn3-job-service",
			"altinn3-outbound-proxy",
			"altinn3-prefill-converter",
			"altinn3-prefill-service",
			"altinn3-subscription-service",
			"altinn3-prefill-api-stat",
		},
		"tannhelse": {
			"tannhelse-iac",
		},
		"tech-coach": {
			"tech-coach-iac",
		},
		"tidsbruk-survey": {
			"tidsbruk-survey-iac",
		},
		"tip-tutorials": {
			"tip-tutorials-iac",
		},
		"transak-bong": {
			"transak-bong-iac",
		},
		"transport-baat": {
			"transport-baat-iac",
		},
		"transport-bil": {
			"transport-bil-iac",
		},
		"transport-fly": {
			"transport-fly-iac",
		},
		"transport-kollekt": {
			"transport-kollekt-iac",
		},
		"transport-kostind": {
			"transport-kostind-iac",
		},
		"transport-kostra": {
			"transport-kostra-iac",
		},
		"transport-lastbil": {
			"transport-lastbil-iac",
		},
		"trygd-person": {
			"trygd-person-iac",
		},
		"uh-tjenester": {
			"uh-tjenester-iac",
		},
		"uh-varer": {
			"uh-varer-iac",
			"uh-varer-api",
			"uh-varer-toll-decl-api",
		},
		"utd-bhgskole": {
			"utd-bhgskole-iac",
		},
		"utd-inter": {
			"utd-inter-iac",
		},
		"utd-kompetanse": {
			"utd-kompetanse-iac",
		},
		"utd-nudb": {
			"utd-nudb-iac",
		},
		"utd-tverr": {
			"utd-tverr-iac",
		},
		"utd-uhfagskole": {
			"utd-uhfagskole-iac",
		},
		"utd-vg": {
			"utd-vg-iac",
		},
		"utslipp-forbruk": {
			"utslipp-forbruk-iac",
		},
		"utslipp-luft": {
			"utslipp-luft-iac",
		},
		"var-avfall": {
			"var-avfall-iac",
		},
		"var-avlop": {
			"var-avlop-iac",
		},
		"var-materialstrom": {
			"var-materialstrom-iac",
		},
		"var-vann-kostra": {
			"var-vann-kostra-iac",
		},
		"vare-tjen-korttid": {
			"vare-tjen-korttid-iac",
		},
		"verdsetting-bolig": {
			"verdsetting-bolig-iac",
		},
		"vof": {
			"vof-iac",
		},
	}
}

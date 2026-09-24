package atlantis

import (
	"context"
	"encoding/base64"
	"fmt"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	googlecreds "golang.org/x/oauth2/google"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

func (r *reconciler) createKubernetesClients(ctx context.Context, clusterResourceName string) (*kubernetes.Clientset, *servingv1.ServingV1Client, error) {
	// Get cluster info
	cluster, err := r.clusterManager.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: clusterResourceName,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("get cluster info: %w", err)
	}

	// Extract server CA cert, base64 encoded. ClusterInfo needs it decoded
	cert := cluster.MasterAuth.ClusterCaCertificate
	ca, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, nil, fmt.Errorf("decode cluster certificate: %w", err)
	}

	endpoint := cluster.ControlPlaneEndpointsConfig.IpEndpointsConfig.GetPublicEndpoint()

	// Get a token source for ADC
	ts, err := googlecreds.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("find default credentials: %w", err)
	}

	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	overrides.ClusterInfo.CertificateAuthorityData = ca
	// Need to specify https, defaults to http
	overrides.ClusterDefaults.Server = "https://" + endpoint

	// Create config for kubernetes clients
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	restCfg, err := cc.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("create rest client config: %w", err)
	}

	// We wrap the underlying HTTP transport with a "middleware" which injects
	// rquests with our ADC
	restCfg.Wrap(transport.TokenSourceWrapTransport(ts.TokenSource))

	k8s, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create kubernetes client: %W", err)
	}

	knsrv, err := servingv1.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create knative serving client: %w", err)
	}

	return k8s, knsrv, nil
}

func (r *reconciler) reconcileKubernetesResources(ctx context.Context, name, namespace string, config *protoapi.AtlantisConfig, repoAllowList []string, log logrus.FieldLogger) error {
	if err := r.reconcileKubernetesServiceAccount(ctx, name, namespace, log); err != nil {
		return fmt.Errorf("reconcile service account: %w", err)
	}

	if err := r.reconcileKubernetesWebhookSecret(ctx, name, namespace, *config.WebhookSecret, log); err != nil {
		return fmt.Errorf("reconcile webhook secret: %w", err)
	}

	repoConfig := defaultRepoConfig
	if len(config.RepoConfig) != 0 {
		repoConfig = string(config.RepoConfig)
	}
	if err := r.reconcileKubernetesReposConfig(ctx, name, namespace, repoConfig, log); err != nil {
		return fmt.Errorf("reconcile repos config: %w", err)
	}

	diskSize := defaultDiskSize
	if config.DiskSize != nil {
		var err error
		diskSize, err = resource.ParseQuantity(*config.DiskSize)
		if err != nil {
			return fmt.Errorf("parse disk size: %w", err)
		}
	}
	if err := r.reconcileKubernetesVolume(ctx, name, namespace, diskSize, log); err != nil {
		return fmt.Errorf("reconcile volume: %w", err)
	}

	if err := r.reconcileKnativeService(ctx, name, namespace, repoAllowList, config, log.WithField("atlantis_subdomain", "knative")); err != nil {
		return fmt.Errorf("reconcile knative service: %w", err)
	}
	return nil
}

func (r *reconciler) reconcileKubernetesServiceAccount(ctx context.Context, name, namespace string, log logrus.FieldLogger) error {
	saClient := r.k8sClient.CoreV1().ServiceAccounts(namespace)
	gcpSaName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, r.config.AtlantisProject)

	wantedAnnotations := map[string]string{
		wiAnnotationKey: gcpSaName,
	}

	sa, err := saClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = saClient.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Annotations: wantedAnnotations,
			},
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if equality.Semantic.DeepDerivative(wantedAnnotations, sa.Annotations) {
		return nil
	}

	if r.config.LogDiffs {
		LogDiff(wantedAnnotations, sa.Annotations, log)
	}

	sa.Annotations = wantedAnnotations
	_, err = saClient.Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesWebhookSecret(ctx context.Context, name, namespace, webhookSecret string, log logrus.FieldLogger) error {
	secretsClient := r.k8sClient.CoreV1().Secrets(namespace)

	wantedData := map[string][]byte{
		webhookSecretKey: []byte(webhookSecret),
	}

	secret, err := secretsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretsClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: wantedData,
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if cmp.Equal(secret.Data, wantedData) {
		return nil
	}

	if r.config.LogDiffs {
		// Hide sensitive secret
		hiddenLive := make(map[string]int, len(secret.Data))
		for k, v := range secret.Data {
			hiddenLive[k] = len(v)
		}
		hiddenWanted := make(map[string]int, len(wantedData))
		for k, v := range secret.Data {
			hiddenWanted[k] = len(v)
		}
		LogDiff(hiddenLive, hiddenWanted, log)
	}

	secret.Data = wantedData
	_, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesReposConfig(ctx context.Context, name, namespace, repoConfig string, log logrus.FieldLogger) error {
	configMapsClient := r.k8sClient.CoreV1().ConfigMaps(namespace)

	wantedData := map[string]string{
		reposYamlKey: repoConfig,
	}

	cm, err := configMapsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = configMapsClient.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: wantedData,
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if cmp.Equal(cm.Data, wantedData) {
		return nil
	}

	if r.config.LogDiffs {
		LogDiff(cm.Data, wantedData, log)
	}

	cm.Data = wantedData
	_, err = configMapsClient.Update(ctx, cm, metav1.UpdateOptions{})

	return err
}

func (r *reconciler) reconcileKubernetesVolume(ctx context.Context, name, namespace string, diskSize resource.Quantity, log logrus.FieldLogger) error {
	pvcClient := r.k8sClient.CoreV1().PersistentVolumeClaims(namespace)
	wantedSpec := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceRequestsStorage: diskSize,
				},
			},
		},
	}

	pvc, err := pvcClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = pvcClient.Create(ctx, wantedSpec, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if equality.Semantic.DeepDerivative(wantedSpec.Spec, pvc.Spec) {
		return nil
	}

	if r.config.LogDiffs {
		LogDiff(pvc.Spec, wantedSpec.Spec, log)
	}

	_, err = pvcClient.Update(ctx, wantedSpec, metav1.UpdateOptions{})
	return err
}

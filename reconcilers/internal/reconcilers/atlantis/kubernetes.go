package atlantis

import (
	"context"
	"encoding/base64"
	"fmt"

	"cloud.google.com/go/container/apiv1/containerpb"
	"github.com/google/go-cmp/cmp"
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

func (r *reconciler) createKubernetesClients(ctx context.Context) (*kubernetes.Clientset, *servingv1.ServingV1Client, error) {
	// Get cluster info
	cluster, err := r.clusterManager.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: r.config.clusterResourceName,
	})
	if err != nil {
		return nil, nil, err
	}

	// Extract server CA cert, base64 encoded. ClusterInfo needs it decoded
	cert := cluster.MasterAuth.ClusterCaCertificate
	ca, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, nil, err
	}

	endpoint := cluster.ControlPlaneEndpointsConfig.IpEndpointsConfig.GetPublicEndpoint()

	// Get a token source for ADC
	ts, err := googlecreds.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	// We wrap the underlying HTTP transport with a "middleware" which injects
	// rquests with our ADC
	restCfg.Wrap(transport.TokenSourceWrapTransport(ts.TokenSource))

	k8s, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, err
	}

	knsrv, err := servingv1.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, err
	}

	return k8s, knsrv, nil
}
func (r *reconciler) reconcileKubernetesResources(ctx context.Context, name, namespace string, config *protoapi.AtlantisConfig, repoAllowList []string) error {
	if err := r.reconcileKubernetesServiceAccount(ctx, name, namespace); err != nil {
		return err
	}

	if err := r.reconcileKubernetesWebhookSecret(ctx, name, namespace, *config.WebhookSecret); err != nil {
		return err
	}

	repoConfig := defaultRepoConfig
	if len(config.RepoConfig) != 0 {
		repoConfig = string(config.RepoConfig)
	}
	if err := r.reconcileKubernetesReposConfig(ctx, name, namespace, repoConfig); err != nil {
		return err
	}

	diskSize := defaultDiskSize
	if config.DiskSize != nil {
		var err error
		diskSize, err = resource.ParseQuantity(*config.DiskSize)
		if err != nil {
			return err
		}
	}
	if err := r.reconcileKubernetesVolume(ctx, name, namespace, diskSize); err != nil {
		return err
	}

	if err := r.reconcileKnativeService(ctx, name, namespace, repoAllowList, config); err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileKubernetesServiceAccount(ctx context.Context, name, namespace string) error {
	saClient := r.k8sClient.CoreV1().ServiceAccounts(namespace)
	gcpSaName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, r.config.atlantisProject)

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

	sa.Annotations = wantedAnnotations
	_, err = saClient.Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesWebhookSecret(ctx context.Context, name, namespace, webhookSecret string) error {
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

	secret.Data = wantedData
	_, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesReposConfig(ctx context.Context, name, namespace, repoConfig string) error {
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

	cm.Data = wantedData
	_, err = configMapsClient.Update(ctx, cm, metav1.UpdateOptions{})

	return err
}

func (r *reconciler) reconcileKubernetesVolume(ctx context.Context, name, namespace string, diskSize resource.Quantity) error {
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

	if equality.Semantic.DeepDerivative(wantedSpec, pvc) {
		return nil
	}

	_, err = pvcClient.Update(ctx, wantedSpec, metav1.UpdateOptions{})
	return err
}

package atlantis

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func (r *reconciler) reconcileKnativeService(ctx context.Context, name string) error {
	env := map[string]string{
		"ATLANTIS_REPO_ALLOWLIST":                        "local.repo_allowlist",
		"ATLANTIS_GH_APP_ID":                             "",
		"ATLANTIS_GH_APP_KEY_FILE":                       "/secret/atlantis-app-key.pem",
		"ATLANTIS_WRITE_GIT_CREDS":                       "true",
		"ATLANTIS_DATA_DIR":                              "/atlantis",
		"ATLANTIS_ATLANTIS_URL":                          "",
		"ATLANTIS_PORT":                                  "4141",
		"TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE": "true",
		"ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY":       "true",
		"ATLANTIS_ENABLE_REGEXP_CMD":                     "true",
		"ATLANTIS_REPO_CONFIG":                           "/config/repos.yaml",
	}
	envVars := make([]corev1.EnvVar, 0, len(env)+1)
	for key, val := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}
	envVars = append(envVars, corev1.EnvVar{
		Name: "ATLANTIS_GH_WEBHOOK_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: webhookSecretKey,
			},
		},
	})

	probe := corev1.Probe{
		PeriodSeconds: 60,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   intstr.FromString("4141"),
				Scheme: corev1.URISchemeHTTP,
			},
		},
	}

	knService := &knv1.Service{
		Spec: knv1.ServiceSpec{
			ConfigurationSpec: knv1.ConfigurationSpec{
				Template: knv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"autoscaling.knative.dev/max-scale":                          "1",
							"autoscaling.knative.dev/scale-to-zero-pod-retention-period": "1h",
						},
					},
					Spec: knv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: name,
							SecurityContext: &corev1.PodSecurityContext{
								FSGroup: new(int64(1000)),
							},
							Containers: []corev1.Container{
								{
									Image: "sdlkfskldfjsdlkfjds TODO",
									Env:   envVars,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "atlantis-data",
											MountPath: "/atlantis",
										},
										{
											Name:      "secret-volume",
											ReadOnly:  true,
											MountPath: "/secret",
										},
										{
											Name:      "config-volume",
											ReadOnly:  true,
											MountPath: "/config",
										},
									},
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 4141,
										},
									},
									LivenessProbe:  &probe,
									ReadinessProbe: &probe,
									Resources: corev1.ResourceRequirements{
										// TODO: investigage resourceXXX vs resource(requests/liits)XXX
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("256Mi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("500m"),
											corev1.ResourceMemory: resource.MustParse("512Mi"), // TODO: get limit from API
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "atlantis-data",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: name,
											ReadOnly:  false,
										},
									},
								},
								{
									Name: "secret-volume",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "dapla-team",
											Items: []corev1.KeyToPath{
												{
													Key:  "gh-key-file",
													Path: "atlantis-app-key.pem",
												},
											},
										},
									},
								},
								{
									Name: "config-volume",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: name,
											},
											Items: []corev1.KeyToPath{
												{
													Key:  "repos.yaml",
													Path: "repos.yaml",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	kns, err := r.knServices.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = r.knServices.Create(ctx, &knv1.Service{}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	return err
}

package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPdPod(cr *tidbclusterv1.Pd) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			/*
				Volumes: []corev1.Volume{
					{
						Name: "pd-pv-storage",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "pd-pv-claim",
							},
						},
					},
				},
			*/
			Containers: []corev1.Container{
				{
					Name:  "pd",
					Image: cr.Spec.Imagename,
					Ports: []corev1.ContainerPort{
						{
							Name:          "client-port",
							ContainerPort: 2379,
						},
						{
							Name:          "peer-port",
							ContainerPort: 2380,
						},
					},
					Args: []string{"--client-urls=http://0.0.0.0:2379", "--advertise-client-urls=http://$(MY_POD_IP):2379",
						"--peer-urls=http://0.0.0.0:2380", "--advertise-peer-urls=http://$(MY_POD_IP):2380"},
					Env: []corev1.EnvVar{
						{
							Name: "MY_POD_IP",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "status.podIP",
								},
							},
						},
					},
					/*
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/default.pd-pd-sample",
								Name:      "pd-pv-storage",
							},
						},
					*/
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

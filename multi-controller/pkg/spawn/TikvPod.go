package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTikvPod(cr *tidbclusterv1.Tikv) *corev1.Pod {
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
			Containers: []corev1.Container{
				{
					Name:  "tikv",
					Image: cr.Spec.Imagename,
					Ports: []corev1.ContainerPort{
						{
							Name:          "tikv-port",
							ContainerPort: 20160,
						},
					},
					Command: []string{"/tikv-server"},
					Args:    []string{"--pd=pd-svc:2379"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

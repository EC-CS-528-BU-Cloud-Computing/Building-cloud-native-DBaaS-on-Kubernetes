package spawn

import (
	dbaasv1alpha1 "dbaas/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPodForCR(cr *dbaasv1alpha1.Pd) *corev1.Pod {
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
					Name:  "pd",
					Image: cr.Spec.Imagename,
					Ports: []corev1.ContainerPort{
						{
							Name:          "pdendpoint",
							ContainerPort: 2379,
						},
						{
							Name:          "monitoring",
							ContainerPort: 2380,
						},
					},
					Args: []string{"--path=127.0.0.1:2379"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

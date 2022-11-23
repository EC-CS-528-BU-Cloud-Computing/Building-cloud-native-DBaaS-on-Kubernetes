package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTidbPod(cr *tidbclusterv1.Tidb) *corev1.Pod {
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
					Name:  "tidb",
					Image: cr.Spec.Imagename,
					Ports: []corev1.ContainerPort{
						{
							Name:          "sql-endpoint",
							ContainerPort: 4000,
						},
						{
							Name:          "monitoring",
							ContainerPort: 10080,
						},
					},
					Command: []string{"/tidb-server"},
					Args:    []string{"--path=pd-svc:2379"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

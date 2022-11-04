package spawn

import (
	tidbopv1alpha1 "tidbOp/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPodForCR(cr *tidbopv1alpha1.Tidbop) *corev1.Pod {
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
							ContainerPort: 20160,
						},
					},
					Args: []string{"--path=http://pd-service"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreatePdSVC(cr *tidbclusterv1.Pd) *corev1.Service {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pd-svc",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       2379,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(2379),
					Name:       "port1",
				},
				{
					Port:       2380,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(2380),
					Name:       "port2",
				},
			},
			//Here it is pretty static, using cr label for svc selector, maybe change in the future?
			Selector: labels,
		},
	}
}

package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTikvSts(cr *tidbclusterv1.Tikv) *appv1.StatefulSet {

	replicaCnt := int32(cr.Spec.Replica)

	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appv1.StatefulSetSpec{
			ServiceName: "tikv-svc",
			Replicas:    &replicaCnt,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "tikv-sts",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cr.Namespace,
					Labels: map[string]string{
						"app": "tikv-sts",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Command: []string{"/tikv-server"},
							Image:   "pingcap/tikv",
							Args:    []string{"--pd=pd-svc:2379", "--addr=$(MY_POD_IP):20160"},
							Name:    "tikv",
							Ports: []v1.ContainerPort{
								{
									Name:          "tikv-port",
									ContainerPort: 20160,
								},
							},
							Env: []v1.EnvVar{
								{
									Name: "MY_POD_IP",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "status.podIP",
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
}

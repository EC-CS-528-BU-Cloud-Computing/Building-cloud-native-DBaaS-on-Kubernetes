package spawn

import (
	tidbclusterv1 "cluster-operator/api/v1"

	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTikvSts(cr *tidbclusterv1.Tikv) *appv1.StatefulSet {
	//labels := map[string]string{
	//	"app": "Tikv",
	//}

	replicaCnt := int32(cr.Spec.Replica)

	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			//Labels:    labels,
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
	/*
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cr.Name,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "tikv-pv-storage",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "tikv-pv-claim",
							},
						},
					},
				},
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
						VolumeMounts: []corev1.VolumeMount{
							{
								MountPath: "/tmp/tikv/store",
								Name:      "tikv-pv-storage",
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyOnFailure,
			},
		}
	*/
}

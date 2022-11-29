package k8s

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetStructFieldValue(t *testing.T) {
	tests := []struct {
		desc        string
		object      any
		path        []string
		updateValue any
		wantObject  any
	}{
		{
			desc: "Update K8s daemonset images",
			object: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "edgedelta",
					Namespace:       "edgedelta",
					UID:             "ef61b019-aaaa-bbbb-cccc-2ddd26e1dbca",
					ResourceVersion: "231953",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Command: []string{"/bin/edgedelta"},
									Env: []v1.EnvVar{
										{
											Name: "ED_API_KEY",
											ValueFrom: &v1.EnvVarSource{
												SecretKeyRef: &v1.SecretKeySelector{
													Key: "ed-api-key",
													LocalObjectReference: v1.LocalObjectReference{
														Name: "ed-api-key",
													},
												},
											},
										},
									},
									Image:           "gcr.io/edgedelta/agent:v0.1.47",
									ImagePullPolicy: v1.PullAlways,
									Name:            "edgedelta",
								},
							},
						},
					},
				},
			},
			path:        []string{"spec", "template", "spec", "containers[0]", "image"},
			updateValue: "gcr.io/edgedelta/agent:v0.1.49",
			wantObject: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "edgedelta",
					Namespace:       "edgedelta",
					UID:             "ef61b019-aaaa-bbbb-cccc-2ddd26e1dbca",
					ResourceVersion: "231953",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Command: []string{"/bin/edgedelta"},
									Env: []v1.EnvVar{
										{
											Name: "ED_API_KEY",
											ValueFrom: &v1.EnvVarSource{
												SecretKeyRef: &v1.SecretKeySelector{
													Key: "ed-api-key",
													LocalObjectReference: v1.LocalObjectReference{
														Name: "ed-api-key",
													},
												},
											},
										},
									},
									Image:           "gcr.io/edgedelta/agent:v0.1.49",
									ImagePullPolicy: v1.PullAlways,
									Name:            "edgedelta",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if err := SetStructFieldValue(tc.object, tc.path, tc.updateValue); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.wantObject, tc.object); diff != "" {
				t.Errorf("Objects mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

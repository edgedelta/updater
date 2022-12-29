package k8s

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompareAndUpdateStructField(t *testing.T) {
	tests := []struct {
		desc        string
		object      any
		path        []string
		updateValue string
		wantObject  any
		wantUpdated bool
	}{
		{
			desc:        "Update K8s daemonset images",
			object:      daemonsetWithImage("gcr.io/edgedelta/agent:v0.1.47"),
			path:        []string{"spec", "template", "spec", "containers[0]", "image"},
			updateValue: "gcr.io/edgedelta/agent:v0.1.49",
			wantUpdated: true,
			wantObject:  daemonsetWithImage("gcr.io/edgedelta/agent:v0.1.49"),
		},
		{
			desc:        "Update K8s daemonset images",
			object:      daemonsetWithImage("gcr.io/edgedelta/agent:v0.1.47"),
			path:        []string{"spec", "template", "spec", "containers[0]", "image"},
			updateValue: "gcr.io/edgedelta/agent:v0.1.47",
			wantUpdated: false,
			wantObject:  daemonsetWithImage("gcr.io/edgedelta/agent:v0.1.47"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			updated, err := CompareAndUpdateStructField(tc.object, tc.path, tc.updateValue)
			if err != nil {
				t.Fatal(err)
			}
			if updated != tc.wantUpdated {
				t.Fatalf("Wanted 'updated' return value as %t, got %t instead", tc.wantUpdated, updated)
			}
			if diff := cmp.Diff(tc.wantObject, tc.object); diff != "" {
				t.Errorf("Objects mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func daemonsetWithImage(image string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
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
							Image:           image,
							ImagePullPolicy: v1.PullAlways,
							Name:            "edgedelta",
						},
					},
				},
			},
		},
	}
}

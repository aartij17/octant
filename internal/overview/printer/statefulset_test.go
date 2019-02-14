package printer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/heptio/developer-dash/internal/cache"
	"github.com/heptio/developer-dash/internal/overview/printer"
	"github.com/heptio/developer-dash/internal/view/component"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_StatefulSetListHandler(t *testing.T) {
	printOptions := printer.Options{
		Cache: cache.NewMemoryCache(),
	}

	labels := map[string]string{
		"foo": "bar",
	}

	now := time.Unix(1547211430, 0)

	object := &appsv1.StatefulSetList{
		Items: []appsv1.StatefulSet{
			{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "web",
					CreationTimestamp: metav1.Time{
						Time: now,
					},
					Labels: labels,
				},
				Status: appsv1.StatefulSetStatus{
					Replicas: 1,
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "myapp",
						},
					},
					Replicas: ptrInt32(3),
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Name:  "nginx",
									Image: "k8s.gcr.io/nginx-slim:0.8",
								},
							},
						},
					},
				},
			},
		},
	}

	got, err := printer.StatefulSetListHandler(object, printOptions)
	require.NoError(t, err)

	cols := component.NewTableCols("Name", "Labels", "Desired", "Current", "Age", "Selector")
	expected := component.NewTable("StatefulSets", cols)
	expected.Add(component.TableRow{
		"Name":     component.NewLink("", "web", "/content/overview/namespace/workloads/stateful-sets/web"),
		"Labels":   component.NewLabels(labels),
		"Desired":  component.NewText("3"),
		"Current":  component.NewText("1"),
		"Age":      component.NewTimestamp(now),
		"Selector": component.NewSelectors([]component.Selector{component.NewLabelSelector("app", "myapp")}),
	})

	assert.Equal(t, expected, got)
}

func Test_StatefulSetStatus(t *testing.T) {
	c := cache.NewMemoryCache()

	labels := map[string]string{
		"app": "myapp",
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "statefulset",
			Namespace: "testing",
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myapp",
				},
			},
		},
	}

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-0",
					Namespace: "testing",
					Labels:    labels,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-1",
					Namespace: "testing",
					Labels:    labels,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-2",
					Namespace: "testing",
					Labels:    labels,
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "random-pod",
					Namespace: "testing",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
		},
	}

	for _, p := range pods.Items {
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(p)
		if err != nil {
			return
		}

		c.Store(&unstructured.Unstructured{Object: u})
	}

	stsc := printer.NewStatefulSetStatus(sts)
	got, err := stsc.Create(c)
	require.NoError(t, err)

	expected := component.NewQuadrant()
	require.NoError(t, expected.Set(component.QuadNW, "Running", "2"))
	require.NoError(t, expected.Set(component.QuadNE, "Waiting", "1"))
	require.NoError(t, expected.Set(component.QuadSW, "Succeeded", "0"))
	require.NoError(t, expected.Set(component.QuadSE, "Failed", "0"))

	assert.Equal(t, expected, got)
}

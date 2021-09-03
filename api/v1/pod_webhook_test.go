package v1

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validatePodTest(pod *v1.Pod, valid bool) {
	ctx := context.Background()

	err := k8sClient.Create(ctx, pod)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "Pod: %v", pod)
	} else {
		Expect(err).To(HaveOccurred(), "Pod: %v", pod)
		statusErr := &k8serrors.StatusError{}
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		expected := pod.Annotations["message"]
		Expect(statusErr.ErrStatus.Message).To(ContainSubstring(expected))
	}
}

var _ = Describe("Pod Webhook", func() {
	Context("validating", func() {
		const (
			PodName      = "test-pod"
			PodNamespace = "default"
			JobName      = "test-job"
		)
		validPod := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      PodName,
				Namespace: PodNamespace,
				Annotations: map[string]string{
					"example-mutating-admission-webhook": "foo",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "nginx",
						Image: "nginx",
					},
				},
			},
		}
		Context("valid", func() {
			It("should create a valid Pod", func() {
				validatePodTest(&validPod, true)
			})
		})

		Context("invalid", func() {
			It("should not create a invalid Pod that have not annotation key 'example-mutating-admission-webhook'", func() {
				invalidPod := validPod.DeepCopy()
				invalidPod.ObjectMeta.Annotations = map[string]string{}
				validatePodTest(&validPod, false)
			})

			It("should not create a invalid Pod that have not annotation 'example-mutating-admission-webhook' values 'foo'", func() {
				invalidPod := validPod.DeepCopy()
				invalidPod.ObjectMeta.Annotations = map[string]string{
					"example-mutating-admission-webhook": "bar",
				}
				validatePodTest(&validPod, false)
			})
		})
	})
})

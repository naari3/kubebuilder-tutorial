package v1

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateTest(cronJob *CronJob, valid bool) {
	ctx := context.Background()

	err := k8sClient.Create(ctx, cronJob)

	if valid {
		Expect(err).NotTo(HaveOccurred(), "CronJob: %v", cronJob)
	} else {
		Expect(err).To(HaveOccurred(), "CronJob: %v", cronJob)
		statusErr := &k8serrors.StatusError{}
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		expected := cronJob.Annotations["message"]
		Expect(statusErr.ErrStatus.Message).To(ContainSubstring(expected))
	}
}

var _ = Describe("CronJob Webhook", func() {
	Context("validating", func() {
		const (
			CronjobName      = "test-cronjob"
			CronjobNamespace = "default"
			JobName          = "test-job"
		)
		validCronJob := CronJob{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "batch.tutorial.kubebuilder.io/v1",
				Kind:       "CronJob",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      CronjobName,
				Namespace: CronjobNamespace,
			},
			Spec: CronJobSpec{
				Schedule: "1 * * * *",
				JobTemplate: batchv1beta1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "test-container",
										Image: "test-image",
									},
								},
								RestartPolicy: v1.RestartPolicyOnFailure,
							},
						},
					},
				},
			},
		}
		It("should create a valid CronJob", func() {
			validateTest(&validCronJob, true)
		})

		invalidCronJob := validCronJob.DeepCopy()
		invalidCronJob.ObjectMeta.Name = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // metadata.name max length is 63

		It("should not create a invalid CronJob", func() {
			validateTest(&validCronJob, false)
		})
	})
})

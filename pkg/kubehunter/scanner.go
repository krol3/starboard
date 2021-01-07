package kubehunter

import (
	"context"
	"fmt"

	"github.com/aquasecurity/starboard/pkg/apis/aquasecurity/v1alpha1"
	"github.com/aquasecurity/starboard/pkg/ext"
	"github.com/aquasecurity/starboard/pkg/kube"
	"github.com/aquasecurity/starboard/pkg/kube/pod"
	"github.com/aquasecurity/starboard/pkg/runner"
	"github.com/aquasecurity/starboard/pkg/scanners"
	"github.com/aquasecurity/starboard/pkg/starboard"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

const (
	kubeHunterContainerName = "kube-hunter"
)

type Config interface {
	GetKubeHunterImageRef() (string, error)
}

type Scanner struct {
	scheme    *runtime.Scheme
	config    Config
	clientset kubernetes.Interface
	ext.IDGenerator
	opts kube.ScannerOpts
	pods *pod.Manager
}

func NewScanner(
	scheme *runtime.Scheme,
	config Config,
	clientset kubernetes.Interface,
	opts kube.ScannerOpts,
) *Scanner {
	return &Scanner{
		scheme:      scheme,
		config:      config,
		clientset:   clientset,
		IDGenerator: ext.NewGoogleUUIDGenerator(),
		opts:        opts,
		pods:        pod.NewPodManager(clientset),
	}
}

func (s *Scanner) Scan(ctx context.Context) (v1alpha1.KubeHunterOutput, error) {
	// 1. Prepare descriptor for the Kubernetes Job which will run kube-hunter
	job, err := s.prepareKubeHunterJob()
	if err != nil {
		return v1alpha1.KubeHunterOutput{}, err
	}

	// 2. Run the prepared Job and wait for its completion or failure
	err = runner.New().Run(ctx, kube.NewRunnableJob(s.scheme, s.clientset, job))
	if err != nil {
		return v1alpha1.KubeHunterOutput{}, fmt.Errorf("running kube-hunter job: %w", err)
	}

	defer func() {
		if !s.opts.DeleteScanJob {
			klog.V(3).Infof("Skipping scan job deletion: %s/%s", job.Namespace, job.Name)
			return
		}
		// 5. Delete the kube-hunter Job
		klog.V(3).Infof("Deleting job: %s/%s", job.Namespace, job.Name)
		background := metav1.DeletePropagationBackground
		_ = s.clientset.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &background,
		})
	}()

	// 3. Get kube-hunter JSON output from the kube-hunter Pod
	klog.V(3).Infof("Getting logs for %s container in job: %s/%s", kubeHunterContainerName,
		job.Namespace, job.Name)
	logsReader, err := s.pods.GetContainerLogsByJob(ctx, job, kubeHunterContainerName)
	if err != nil {
		return v1alpha1.KubeHunterOutput{}, fmt.Errorf("getting logs: %w", err)
	}
	defer func() {
		_ = logsReader.Close()
	}()

	// 4. Parse the KubeHuberOutput from the logs Reader
	return OutputFrom(s.config, logsReader)
}

func (s *Scanner) prepareKubeHunterJob() (*batchv1.Job, error) {
	imageRef, err := s.config.GetKubeHunterImageRef()
	if err != nil {
		return nil, err
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.GenerateID(),
			Namespace: starboard.NamespaceName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          pointer.Int32Ptr(0),
			Completions:           pointer.Int32Ptr(1),
			ActiveDeadlineSeconds: scanners.GetActiveDeadlineSeconds(s.opts.ScanJobTimeout),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: starboard.ServiceAccountName,
					RestartPolicy:      corev1.RestartPolicyNever,
					HostPID:            true,
					Affinity:           starboard.DefaultAffinity(),
					Containers: []corev1.Container{
						{
							Name:                     kubeHunterContainerName,
							Image:                    imageRef,
							ImagePullPolicy:          corev1.PullIfNotPresent,
							TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
							Args:                     []string{"--pod", "--report", "json", "--log", "warn"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("300m"),
									corev1.ResourceMemory: resource.MustParse("400M"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

package vulnerabilityreport

import (
	"context"
	"fmt"

	"github.com/aquasecurity/starboard/pkg/docker"
	"github.com/aquasecurity/starboard/pkg/kube"
	"github.com/aquasecurity/starboard/pkg/runner"
	"github.com/aquasecurity/starboard/pkg/starboard"
	//"github.com/krol3/starboard/pkg/apis/wgpolicyk8s.io/v1alpha2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Scanner is a template for running static vulnerability scanners that implement
// the Plugin interface.
type Scanner struct {
	scheme         *runtime.Scheme
	clientset      kubernetes.Interface
	plugin         Plugin
	pluginContext  starboard.PluginContext
	objectResolver *kube.ObjectResolver
	logsReader     kube.LogsReader
	config         starboard.ConfigData
	opts           kube.ScannerOpts
	secretsReader  kube.SecretsReader
}

// NewScanner constructs a new static vulnerability Scanner with the specified
// Plugin that knows how to perform the actual scanning,
// which is performed by running a Kubernetes job, and knows how to convert logs
// to instances of v1alpha2.PolicyReport.
func NewScanner(
	clientset kubernetes.Interface,
	client client.Client,
	plugin Plugin,
	pluginContext starboard.PluginContext,
	config starboard.ConfigData,
	opts kube.ScannerOpts,
) *Scanner {
	return &Scanner{
		scheme:         client.Scheme(),
		clientset:      clientset,
		opts:           opts,
		plugin:         plugin,
		pluginContext:  pluginContext,
		objectResolver: &kube.ObjectResolver{Client: client},
		logsReader:     kube.NewLogsReader(clientset),
		config:         config,
		secretsReader:  kube.NewSecretsReader(client),
	}
}

// Scan creates a Kubernetes job to scan the specified workload. The pod created
// by the scan job has template contributed by the Plugin.
// It is a blocking method that watches the status of the job until it succeeds
// or fails. When succeeded it parses container logs and coverts the output
// to instances of v1alpha1.VulnerabilityReport by delegating such transformation
// logic also to the Plugin.
func (s *Scanner) Scan(ctx context.Context, workload kube.Object) ([]v1alpha2.PolicyReport, error) {

	klog.V(3).Infof("Getting Pod template for workload: %v", workload)

	owner, err := s.objectResolver.GetObjectFromPartialObject(ctx, workload)
	if err != nil {
		return nil, fmt.Errorf("resolving object: %w", err)
	}
	spec, err := kube.GetPodSpec(owner)
	if err != nil {
		return nil, fmt.Errorf("getting Pod template: %w", err)
	}

	klog.V(3).Infof("Scanning with options: %+v", s.opts)

	credentials, err := s.getCredentials(ctx, spec, workload.Namespace)
	if err != nil {
		return nil, err
	}

	job, secrets, err := s.prepareScanJob(owner, spec, credentials)
	if err != nil {
		return nil, fmt.Errorf("preparing scan job: %w", err)
	}

	err = runner.New().Run(ctx, kube.NewRunnableJob(s.scheme, s.clientset, job, secrets...))
	if err != nil {
		return nil, fmt.Errorf("running scan job: %w", err)
	}

	defer func() {
		if !s.opts.DeleteScanJob {
			klog.V(3).Infof("Skipping scan job deletion: %s/%s", job.Namespace, job.Name)
			return
		}
		klog.V(3).Infof("Deleting scan job: %s/%s", job.Namespace, job.Name)
		background := metav1.DeletePropagationBackground
		_ = s.clientset.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &background,
		})
	}()

	klog.V(3).Infof("Scan job completed: %s/%s", job.Namespace, job.Name)

	return s.getVulnerabilityReportsByScanJob(ctx, job, owner)
}

func (s *Scanner) getCredentials(ctx context.Context, spec corev1.PodSpec, ns string) (map[string]docker.Auth, error) {
	imagePullSecrets, err := s.secretsReader.ListImagePullSecretsByPodSpec(ctx, spec, ns)
	if err != nil {
		return nil, err
	}
	return kube.MapContainerNamesToDockerAuths(kube.GetContainerImagesFromPodSpec(spec), imagePullSecrets)
}

func (s *Scanner) prepareScanJob(workload client.Object, spec corev1.PodSpec, credentials map[string]docker.Auth) (*batchv1.Job, []*corev1.Secret, error) {
	templateSpec, secrets, err := s.plugin.GetScanJobSpec(s.pluginContext, spec, credentials)
	if err != nil {
		return nil, nil, err
	}

	scanJobTolerations, err := s.config.GetScanJobTolerations()
	if err != nil {
		return nil, nil, err
	}
	templateSpec.Tolerations = append(templateSpec.Tolerations, scanJobTolerations...)

	templateSpec.ServiceAccountName = starboard.ServiceAccountName

	containerImagesAsJSON, err := kube.GetContainerImagesFromPodSpec(spec).AsJSON()
	if err != nil {
		return nil, nil, err
	}

	scanJobAnnotations, err := s.config.GetScanJobAnnotations()
	if err != nil {
		return nil, nil, err
	}

	podSpecHash := kube.ComputeHash(spec)

	labels := map[string]string{
		starboard.LabelResourceKind:               workload.GetObjectKind().GroupVersionKind().Kind,
		starboard.LabelResourceName:               workload.GetName(),
		starboard.LabelResourceNamespace:          workload.GetNamespace(),
		starboard.LabelPodSpecHash:                podSpecHash,
		starboard.LabelK8SAppManagedBy:            starboard.AppStarboard,
		starboard.LabelVulnerabilityReportScanner: "true",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetScanJobName(workload),
			Namespace: starboard.NamespaceName,
			Labels:    labels,
			Annotations: map[string]string{
				starboard.AnnotationContainerImages: containerImagesAsJSON,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          pointer.Int32Ptr(0),
			Completions:           pointer.Int32Ptr(1),
			ActiveDeadlineSeconds: kube.GetActiveDeadlineSeconds(s.opts.ScanJobTimeout),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: scanJobAnnotations,
				},
				Spec: templateSpec,
			},
		},
	}, secrets, nil
}

// TODO To make this method look the same as the one used by the operator we
// should resolve the owner based on labels set on the given job instead of
// passing owner directly. The goal is for CLI and operator to create jobs
// with the same struct and set of labels to reuse code responsible for parsing
// v1alpha1.VulnerabilityReport instances.
func (s *Scanner) getVulnerabilityReportsByScanJob(ctx context.Context, job *batchv1.Job, owner client.Object) ([]v1alpha2.PolicyReport, error) {
	var reports []v1alpha2.PolicyReport

	containerImages, err := kube.GetContainerImagesFromJob(job)

	if err != nil {
		return nil, fmt.Errorf("getting container images: %w", err)
	}

	podSpecHash, ok := job.Labels[starboard.LabelPodSpecHash]
	if !ok {
		return nil, fmt.Errorf("expected label %s not set", starboard.LabelPodSpecHash)
	}

	for containerName, containerImage := range containerImages {
		klog.V(3).Infof("Getting logs for %s container in job: %s/%s", containerName, job.Namespace, job.Name)
		logsStream, err := s.logsReader.GetLogsByJobAndContainerName(ctx, job, containerName)
		if err != nil {
			return nil, err
		}
		result, err := s.plugin.ParseVulnerabilityReportData(s.pluginContext, containerImage, logsStream)
		if err != nil {
			return nil, err
		}

		_ = logsStream.Close()

		report, err := NewReportBuilder(s.scheme).
			Controller(owner).
			Container(containerName).
			Data(result).
			PodSpecHash(podSpecHash).
			Get()
		if err != nil {
			return nil, err
		}

		reports = append(reports, report)

	}
	return reports, nil
}

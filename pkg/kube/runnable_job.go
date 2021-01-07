package kube

import (
	"context"
	"fmt"
	"time"

	"github.com/aquasecurity/starboard/pkg/runner"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	defaultResyncDuration = 30 * time.Minute
)

type runnableJob struct {
	scheme    *runtime.Scheme
	clientset kubernetes.Interface

	job     *batchv1.Job     // job to be run
	secrets []*corev1.Secret // secrets that the job references
}

// NewRunnableJob constructs a new Runnable task defined as Kubernetes
// job configuration and secrets that it references.
func NewRunnableJob(
	scheme *runtime.Scheme,
	clientset kubernetes.Interface,
	job *batchv1.Job,
	secrets ...*corev1.Secret,
) runner.Runnable {
	return &runnableJob{
		scheme:    scheme,
		job:       job,
		secrets:   secrets,
		clientset: clientset,
	}
}

// Run runs synchronously the task as Kubernetes job.
// It creates Kubernetes job and secrets provided as constructor parameters.
// This method blocks and waits for the job completion or failure.
// For each secret it also sets the owner reference that points to the job
// so when the job is deleted secrets are garbage collected.
func (r *runnableJob) Run(ctx context.Context) error {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		r.clientset,
		defaultResyncDuration,
		informers.WithNamespace(r.job.Namespace),
	)
	jobInformer := informerFactory.Batch().V1().Jobs()

	var err error

	for i, secret := range r.secrets {
		klog.V(3).Infof("Creating secret %q", r.job.Namespace+"/"+secret.Name)
		r.secrets[i], err = r.clientset.CoreV1().Secrets(r.job.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating secret: %w", err)
		}
	}

	klog.V(3).Infof("Creating job %q", r.job.Namespace+"/"+r.job.Name)
	r.job, err = r.clientset.BatchV1().Jobs(r.job.Namespace).Create(ctx, r.job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}

	for i, secret := range r.secrets {
		klog.V(3).Infof("Setting owner reference secret %q -> job %q", r.job.Namespace+"/"+secret.Name, r.job.Namespace+"/"+r.job.Name)
		err = controllerutil.SetOwnerReference(r.job, secret, r.scheme)
		if err != nil {
			return fmt.Errorf("setting owner reference: %w", err)
		}
		klog.V(3).Infof("Updating secret %q", r.job.Namespace+"/"+secret.Name)
		r.secrets[i], err = r.clientset.CoreV1().Secrets(r.job.Namespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("updating secret: %w", err)
		}
	}

	complete := make(chan error)

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newJob, ok := newObj.(*batchv1.Job)
			if !ok {
				return
			}
			if r.job.UID != newJob.UID {
				return
			}
			if len(newJob.Status.Conditions) == 0 {
				return
			}
			switch condition := newJob.Status.Conditions[0]; condition.Type {
			case batchv1.JobComplete:
				klog.V(3).Infof("Stopping runnable job on task completion with status: %s", batchv1.JobComplete)
				complete <- nil
			case batchv1.JobFailed:
				klog.V(3).Infof("Stopping runnable job on task failure with status: %s", batchv1.JobFailed)
				complete <- fmt.Errorf("job failed: %s: %s", condition.Reason, condition.Message)
			}
		},
	})

	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	err = <-complete
	return err
}

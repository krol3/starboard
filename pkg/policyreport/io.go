package policyreport

import (
	"context"
	"fmt"

	"github.com/aquasecurity/starboard/pkg/apis/aquasecurity/v1alpha1"
	"github.com/aquasecurity/starboard/pkg/kube"
	"github.com/aquasecurity/starboard/pkg/starboard"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Writer is the interface that wraps the basic Write method.
//
// Write creates or updates the given slice of v1alpha1.VulnerabilityReport
// instances.
type Writer interface {
	Write(context.Context, []v1alpha1.VulnerabilityReport) error
}

// Reader is the interface that wraps methods for finding v1alpha1.VulnerabilityReport objects.
//
// FindByOwner returns the slice of v1alpha1.VulnerabilityReport instances
// owned by the given kube.Object or an empty slice if the reports are not found.
//
// FindByOwnerInHierarchy is similar to FindByOwner except it tries to lookup
// v1alpha1.VulnerabilityReport objects owned by related Kubernetes objects.
// For example, if the given owner is a Deployment, but reports are owned by the
// active ReplicaSet (current revision) this method will return the reports.
type Reader interface {
	FindByOwner(context.Context, kube.Object) ([]v1alpha1.VulnerabilityReport, error)
	FindByOwnerInHierarchy(ctx context.Context, object kube.Object) ([]v1alpha1.VulnerabilityReport, error)
}

type ReadWriter interface {
	Reader
	Writer
}

type readWriter struct {
	*kube.ObjectResolver
}

// NewReadWriter constructs a new ReadWriter which is using the client package
// provided by the controller-runtime libraries for interacting with the
// Kubernetes API server.
func NewReadWriter(client client.Client) ReadWriter {
	return &readWriter{
		ObjectResolver: &kube.ObjectResolver{Client: client},
	}
}

func (r *readWriter) Write(ctx context.Context, reports []v1alpha1.VulnerabilityReport) error {
	for _, report := range reports {
		err := r.createOrUpdate(ctx, report)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *readWriter) createOrUpdate(ctx context.Context, report v1alpha1.VulnerabilityReport) error {
	var existing v1alpha1.VulnerabilityReport
	err := r.Get(ctx, types.NamespacedName{
		Name:      report.Name,
		Namespace: report.Namespace,
	}, &existing)

	if err == nil {
		copied := existing.DeepCopy()
		copied.Labels = report.Labels
		copied.Report = report.Report

		return r.Update(ctx, copied)
	}

	if errors.IsNotFound(err) {
		return r.Create(ctx, &report)
	}

	return err
}

func (r *readWriter) FindByOwner(ctx context.Context, owner kube.Object) ([]v1alpha1.VulnerabilityReport, error) {
	var list v1alpha1.VulnerabilityReportList

	err := r.List(ctx, &list, client.MatchingLabels{
		starboard.LabelResourceKind:      string(owner.Kind),
		starboard.LabelResourceNamespace: owner.Namespace,
		starboard.LabelResourceName:      owner.Name,
	}, client.InNamespace(owner.Namespace))
	if err != nil {
		return nil, err
	}

	return list.DeepCopy().Items, nil
}

func (r *readWriter) FindByOwnerInHierarchy(ctx context.Context, owner kube.Object) ([]v1alpha1.VulnerabilityReport, error) {
	reports, err := r.FindByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}

	// no reports found for provided owner, look for reports in related replicaset
	if len(reports) == 0 && (owner.Kind == kube.KindDeployment || owner.Kind == kube.KindPod) {
		rsName, err := r.GetRelatedReplicasetName(ctx, owner)
		if err != nil {
			return nil, fmt.Errorf("getting replicaset related to %s/%s: %w", owner.Kind, owner.Name, err)
		}
		reports, err = r.FindByOwner(ctx, kube.Object{
			Kind:      kube.KindReplicaSet,
			Name:      rsName,
			Namespace: owner.Namespace,
		})
		if err != nil {
			return nil, err
		}
	}

	return reports, nil
}

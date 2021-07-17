package vulnerabilityreport

import (
	"fmt"
	"strings"

	"github.com/aquasecurity/starboard/pkg/apis/aquasecurity/v1alpha1"
	"github.com/aquasecurity/starboard/pkg/kube"
	"github.com/aquasecurity/starboard/pkg/starboard"
	//"github.com/krol3/starboard/pkg/apis/wgpolicy.io/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetScanJobName(obj client.Object) string {
	return fmt.Sprintf("scan-vulnerabilityreport-%s", kube.ComputeHash(kube.Object{
		Kind:      kube.Kind(obj.GetObjectKind().GroupVersionKind().Kind),
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}))
}

type ReportBuilder struct {
	scheme    *runtime.Scheme
	owner     metav1.Object
	container string
	hash      string
	data      v1alpha2.PolicyReport
}

func NewReportBuilder(scheme *runtime.Scheme) *ReportBuilder {
	return &ReportBuilder{
		scheme: scheme,
	}
}

func (b *ReportBuilder) Controller(owner metav1.Object) *ReportBuilder {
	b.owner = owner
	return b
}

func (b *ReportBuilder) Container(name string) *ReportBuilder {
	b.container = name
	return b
}

func (b *ReportBuilder) PodSpecHash(hash string) *ReportBuilder {
	b.hash = hash
	return b
}

func (b *ReportBuilder) Data(data v1alpha1.VulnerabilityReportData) *ReportBuilder {
	b.data = data
	return b
}

func (b *ReportBuilder) reportName() (string, error) {
	kind, err := kube.KindForObject(b.owner, b.scheme)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", strings.ToLower(kind),
		b.owner.GetName(), b.container), nil
}

func (b *ReportBuilder) Get() (v1alpha1.VulnerabilityReport, error) {
	kind, err := kube.KindForObject(b.owner, b.scheme)
	if err != nil {
		return v1alpha1.VulnerabilityReport{}, fmt.Errorf("getting kind for object: %w", err)
	}

	labels := map[string]string{
		starboard.LabelResourceKind:      kind,
		starboard.LabelResourceName:      b.owner.GetName(),
		starboard.LabelResourceNamespace: b.owner.GetNamespace(),
		starboard.LabelContainerName:     b.container,
	}

	if b.hash != "" {
		labels[starboard.LabelPodSpecHash] = b.hash
	}

	reportName, err := b.reportName()
	if err != nil {
		return v1alpha2.PolicyReport{}, err
	}

	report := v1alpha2.PolicyReport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reportName,
			Namespace: b.owner.GetNamespace(),
			Labels:    labels,
		},
		Report: b.data,
	}
	err = controllerutil.SetControllerReference(b.owner, &report, b.scheme)
	if err != nil {
		return v1alpha2.PolicyReport{}, fmt.Errorf("setting controller reference: %w", err)
	}
	// The OwnerReferencesPermissionsEnforcement admission controller protects the
	// access to metadata.ownerReferences[x].blockOwnerDeletion of an object, so
	// that only users with "update" permission to the finalizers subresource of the
	// referenced owner can change it.
	// We set metadata.ownerReferences[x].blockOwnerDeletion to false so that
	// additional RBAC permissions are not required when the OwnerReferencesPermissionsEnforcement
	// is enabled.
	// See https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#ownerreferencespermissionenforcement
	report.OwnerReferences[0].BlockOwnerDeletion = pointer.BoolPtr(false)
	return report, nil
}

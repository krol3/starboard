package policyreport

import (
	"io"

	"github.com/aquasecurity/starboard/pkg/docker"
	"github.com/aquasecurity/starboard/pkg/starboard"
	"github.com/krol3/starboard/pkg/apis/wgpolicyk8s.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

// Plugin defines the interface between Starboard and static vulnerability
// scanners.
type Plugin interface {

	// Init is a callback to initialize this plugin, e.g. ensure the default
	// configuration.
	Init(ctx starboard.PluginContext) error

	// GetScanJobSpec describes the pod that will be created by Starboard when
	// it schedules a Kubernetes job to scan the workload with the specified
	// descriptor.
	// The second argument maps container names to Docker registry credentials,
	// which can be passed to the scanner as environment variables with values
	// set from returned secrets.
	GetScanJobSpec(ctx starboard.PluginContext, spec corev1.PodSpec, credentials map[string]docker.Auth) (
		corev1.PodSpec, []*corev1.Secret, error)

	// ParseVulnerabilityReportData is a callback to parse and convert logs of
	// the pod controlled by the scan job to v1alpha1.VulnerabilityScanResult.
	ParseVulnerabilityReportData(ctx starboard.PluginContext, imageRef string, logsReader io.ReadCloser) (
		v1alpha2.PolicyReport, error)
}

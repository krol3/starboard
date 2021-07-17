package policyreport

import (
	"sort"

	"github.com/aquasecurity/starboard/pkg/apis/aquasecurity/v1alpha1"
)

type Vulnerabilities []v1alpha1.Vulnerability

func (s Vulnerabilities) Len() int { return len(s) }

func (s Vulnerabilities) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// BySeverity implements sort.Interface by providing Less and using the
// Vulnerabilities.Len and Vulnerabilities.Swap methods of the embedded
// Vulnerabilities value.
type BySeverity struct{ Vulnerabilities }

var severityOrder = map[v1alpha1.Severity]int{
	v1alpha1.SeverityCritical: 0,
	v1alpha1.SeverityHigh:     1,
	v1alpha1.SeverityMedium:   2,
	v1alpha1.SeverityLow:      3,
	v1alpha1.SeverityUnknown:  4,
}

func (s BySeverity) Less(i, j int) bool {
	return severityOrder[s.Vulnerabilities[i].Severity] < severityOrder[s.Vulnerabilities[j].Severity]
}

type LessFunc func(p1, p2 *v1alpha1.VulnerabilityReport) bool

// multiSorter implements the Sort interface, sorting the reports within.
type multiSorter struct {
	reports []v1alpha1.VulnerabilityReport
	less    []LessFunc
}

// SortDesc sorts the argument slice according to the LessFunc functions passed to OrderedBy.
func (ms *multiSorter) SortDesc(reports []v1alpha1.VulnerabilityReport) {
	ms.reports = reports
	sort.Stable(sort.Reverse(ms))
}

// OrderedBy returns a Sorter that sorts using the LessFunc functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...LessFunc) *multiSorter {
	return &multiSorter{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
	return len(ms.reports)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
	ms.reports[i], ms.reports[j] = ms.reports[j], ms.reports[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that discriminates between
// the two items (one is less than the other). Note that it can call the
// less functions twice per call. We could change the functions to return
// -1, 0, 1 and reduce the number of calls for greater efficiency: an
// exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
	p, q := &ms.reports[i], &ms.reports[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}

var (
	SummaryCount = []LessFunc{
		func(r1, r2 *v1alpha1.VulnerabilityReport) bool {
			return r1.Report.Summary.CriticalCount < r2.Report.Summary.CriticalCount
		}, func(r1, r2 *v1alpha1.VulnerabilityReport) bool {
			return r1.Report.Summary.HighCount < r2.Report.Summary.HighCount
		}, func(r1, r2 *v1alpha1.VulnerabilityReport) bool {
			return r1.Report.Summary.MediumCount < r2.Report.Summary.MediumCount
		}, func(r1, r2 *v1alpha1.VulnerabilityReport) bool {
			return r1.Report.Summary.LowCount < r2.Report.Summary.LowCount
		}, func(r1, r2 *v1alpha1.VulnerabilityReport) bool {
			return r1.Report.Summary.UnknownCount < r2.Report.Summary.UnknownCount
		}}
)

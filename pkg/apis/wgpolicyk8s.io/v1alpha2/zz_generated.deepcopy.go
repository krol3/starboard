// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha2

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPolicyReport) DeepCopyInto(out *ClusterPolicyReport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Scope != nil {
		in, out := &in.Scope, &out.Scope
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.ScopeSelector != nil {
		in, out := &in.ScopeSelector, &out.ScopeSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	out.Summary = in.Summary
	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]*PolicyReportResult, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PolicyReportResult)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPolicyReport.
func (in *ClusterPolicyReport) DeepCopy() *ClusterPolicyReport {
	if in == nil {
		return nil
	}
	out := new(ClusterPolicyReport)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPolicyReport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterPolicyReportList) DeepCopyInto(out *ClusterPolicyReportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterPolicyReport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterPolicyReportList.
func (in *ClusterPolicyReportList) DeepCopy() *ClusterPolicyReportList {
	if in == nil {
		return nil
	}
	out := new(ClusterPolicyReportList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterPolicyReportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyReport) DeepCopyInto(out *PolicyReport) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Scope != nil {
		in, out := &in.Scope, &out.Scope
		*out = new(v1.ObjectReference)
		**out = **in
	}
	if in.ScopeSelector != nil {
		in, out := &in.ScopeSelector, &out.ScopeSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	out.Summary = in.Summary
	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]*PolicyReportResult, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(PolicyReportResult)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyReport.
func (in *PolicyReport) DeepCopy() *PolicyReport {
	if in == nil {
		return nil
	}
	out := new(PolicyReport)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolicyReport) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyReportList) DeepCopyInto(out *PolicyReportList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolicyReport, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyReportList.
func (in *PolicyReportList) DeepCopy() *PolicyReportList {
	if in == nil {
		return nil
	}
	out := new(PolicyReportList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolicyReportList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyReportResult) DeepCopyInto(out *PolicyReportResult) {
	*out = *in
	out.Timestamp = in.Timestamp
	if in.Subjects != nil {
		in, out := &in.Subjects, &out.Subjects
		*out = make([]*v1.ObjectReference, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.ObjectReference)
				**out = **in
			}
		}
	}
	if in.SubjectSelector != nil {
		in, out := &in.SubjectSelector, &out.SubjectSelector
		*out = new(metav1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyReportResult.
func (in *PolicyReportResult) DeepCopy() *PolicyReportResult {
	if in == nil {
		return nil
	}
	out := new(PolicyReportResult)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyReportSummary) DeepCopyInto(out *PolicyReportSummary) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyReportSummary.
func (in *PolicyReportSummary) DeepCopy() *PolicyReportSummary {
	if in == nil {
		return nil
	}
	out := new(PolicyReportSummary)
	in.DeepCopyInto(out)
	return out
}

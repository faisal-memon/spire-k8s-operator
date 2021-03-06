// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.ClusterSpiffeId": schema_pkg_apis_spiffeid_v1alpha1_ClusterSpiffeId(ref),
		"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeId":        schema_pkg_apis_spiffeid_v1alpha1_SpiffeId(ref),
		"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdSpec":    schema_pkg_apis_spiffeid_v1alpha1_SpiffeIdSpec(ref),
		"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdStatus":  schema_pkg_apis_spiffeid_v1alpha1_SpiffeIdStatus(ref),
	}
}

func schema_pkg_apis_spiffeid_v1alpha1_ClusterSpiffeId(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ClusterSpiffeId is the Schema for the spiffeids API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdSpec", "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_spiffeid_v1alpha1_SpiffeId(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpiffeId is the Schema for the spiffeids API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdSpec", "github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.SpiffeIdStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_spiffeid_v1alpha1_SpiffeIdSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpiffeIdSpec defines the desired state of SpiffeId",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"spiffeId": {
						SchemaProps: spec.SchemaProps{
							Description: "The Spiffe ID to create",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"selector": {
						SchemaProps: spec.SchemaProps{
							Description: "Selectors to match for this ID",
							Ref:         ref("github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.Selector"),
						},
					},
				},
				Required: []string{"spiffeId", "selector"},
			},
		},
		Dependencies: []string{
			"github.com/transferwise/spire-k8s-operator/pkg/apis/spiffeid/v1alpha1.Selector"},
	}
}

func schema_pkg_apis_spiffeid_v1alpha1_SpiffeIdStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SpiffeIdStatus defines the observed state of SpiffeId",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"entryId": {
						SchemaProps: spec.SchemaProps{
							Description: "The spire Entry ID created for this Spiffe ID",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"entryId"},
			},
		},
	}
}

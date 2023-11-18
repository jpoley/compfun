package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"AddTwoBuckets": {
			reason: "The Function should add two buckets to the desired composed resources",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							// MustStructJSON is a handy way to provide mock
							// resources.
							Resource: resource.MustStructJSON(`{
								"apiVersion": "example.crossplane.io/v1alpha1",
								"kind": "XBuckets",
								"metadata": {
									"name": "test"
								},
								"spec": {
									"region": "us-east-2",
									"names": [
										"test-bucket-a",
										"test-bucket-b"
									]
								}
							}`),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Ttl: durationpb.New(60 * time.Second)},
					Desired: &fnv1beta1.State{
						Resources: map[string]*fnv1beta1.Resource{
							"xbuckets-test-bucket-a": {Resource: resource.MustStructJSON(`{
								"apiVersion": "s3.aws.upbound.io/v1beta1",
								"kind": "Bucket",
								"metadata": {
									"annotations": {
										"crossplane.io/external-name": "test-bucket-a"
									}
								},
								"spec": {
									"forProvider": {
										"region": "us-east-2"
									}
								}
							}`)},
							"xbuckets-test-bucket-b": {Resource: resource.MustStructJSON(`{
								"apiVersion": "s3.aws.upbound.io/v1beta1",
								"kind": "Bucket",
								"metadata": {
									"annotations": {
										"crossplane.io/external-name": "test-bucket-b"
									}
								},
								"spec": {
									"forProvider": {
										"region": "us-east-2"
									}
								}
							}`)},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}

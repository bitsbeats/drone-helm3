package kube

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKube(t *testing.T) {
	kubeconfig := "/tmp/drone-helm3.tmp"
	tests := []struct {
		options []Option
		want    string
		err     error
	}{
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithToken("token"),
				WithNamespace("myapp"),
			},
			want: `
apiVersion: v1
kind: Config

current-context: "helm"
preferences: {}

clusters:
  - name: helm
    cluster:
      server: https://example.com

users:
- name: helm
  user:
    token: token

contexts:
  - name: helm
    context:
      cluster: helm
      namespace: myapp
      user: helm
`,
		},
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithCertificate("CERTDATA"),
				WithToken("token"),
				WithNamespace("myapp"),
			},
			want: `
apiVersion: v1
kind: Config

current-context: "helm"
preferences: {}

clusters:
  - name: helm
    cluster:
      server: https://example.com
      certificate-authority-data: CERTDATA

users:
- name: helm
  user:
    token: token

contexts:
  - name: helm
    context:
      cluster: helm
      namespace: myapp
      user: helm
`,
		},
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithCertificate("CERTDATA"),
				WithEKSCluster("eks_cluster"),
				WithEKSRoleARN("my-eks-role-arn"),
				WithNamespace("myapp"),
			},
			want: `
apiVersion: v1
kind: Config

current-context: "helm"
preferences: {}

clusters:
  - name: helm
    cluster:
      server: https://example.com
      certificate-authority-data: CERTDATA

users:
- name: helm
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws-iam-authenticator
      args:
        - "token"
        - "-i"
        - "eks_cluster"
        - "-r"
        - "my-eks-role-arn"

contexts:
  - name: helm
    context:
      cluster: helm
      namespace: myapp
      user: helm
`,
		},		
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithSkipTLS(true),
				WithToken("token"),
				WithNamespace("myapp"),
			},
			want: `
apiVersion: v1
kind: Config

current-context: "helm"
preferences: {}

clusters:
  - name: helm
    cluster:
      server: https://example.com
      insecure-skip-tls-verify: true

users:
- name: helm
  user:
    token: token

contexts:
  - name: helm
    context:
      cluster: helm
      namespace: myapp
      user: helm
`,
		},

		{
			options: []Option{
				WithApiServer("https://example.com"),
				WithToken("token"),
				WithNamespace("myapp"),
			},
			err: fmt.Errorf("no path to kubeconfig provided"),
		},
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithToken("token"),
				WithNamespace("myapp"),
			},
			err: fmt.Errorf("no kubernetes api server provided"),
		},
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithToken("token"),
				WithEKSCluster("eks_cluster"),
				WithNamespace("myapp"),
			},
			err: fmt.Errorf("token cannot be used simultaneously with eksCluster"),
		},
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithNamespace("myapp"),
			},
			err: fmt.Errorf("no kubernetes token provided"),
		},		
		{
			options: []Option{
				WithConfig(kubeconfig),
				WithApiServer("https://example.com"),
				WithToken("token"),
			},
			err: fmt.Errorf("no namespace provided"),
		},
	}
	for _, test := range tests {
		_ = os.Remove(kubeconfig)
		err := CreateKubeConfig(test.options...)
		if !errEq(err, test.err) {
			t.Fatalf("unable to create kubeconfig: %s", err)
		} else if err != nil {
			continue
		}
		data, err := ioutil.ReadFile("/tmp/drone-helm3.tmp")
		if err != nil {
			t.Fatalf("unable to reade kubeconfig: %s", err)
		}
		got := string(data)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Fatalf(diff)
		}
	}
}

func errEq(a error, b error) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

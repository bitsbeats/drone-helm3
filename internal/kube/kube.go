package kube

import (
	"fmt"
	"os"
	"text/template"
)

type (
	kubeConfig struct {
		Config      string
		ApiServer   string
		Token       string
		Certificate string
		SkipTLS     bool
		Namespace   string
		EKSCluster  string
		EKSRoleARN  string
	}

	Option func(*kubeConfig)
)

var tmpl = template.Must(template.New("kubeconfig").Parse(`
apiVersion: v1
kind: Config

current-context: "helm"
preferences: {}

clusters:
  - name: helm
    cluster:
      server: {{ .ApiServer }}
      {{- if eq .SkipTLS true }}
      insecure-skip-tls-verify: true
      {{- else if not (eq .Certificate "") }}
      certificate-authority-data: {{ .Certificate }}
      {{- end}}

users:
- name: helm
  user:
{{- if .Token }}
    token: {{ .Token }}
{{- else if .EKSCluster }}
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws-iam-authenticator
      args:
        - "token"
        - "-i"
        - "{{ .EKSCluster }}"
    {{- if .EKSRoleARN }}
        - "-r"
        - "{{ .EKSRoleARN }}"
    {{- end }}
{{- end }}

contexts:
  - name: helm
    context:
      cluster: helm
      namespace: {{ .Namespace }}
      user: helm
`))

func WithConfig(config string) Option {
	return func(k *kubeConfig) {
		k.Config = config
	}
}

func WithApiServer(apiServer string) Option {
	return func(k *kubeConfig) {
		k.ApiServer = apiServer
	}
}

func WithToken(token string) Option {
	return func(k *kubeConfig) {
		k.Token = token
	}
}

func WithEKSCluster(eksCluster string) Option {
	return func(k *kubeConfig) {
		k.EKSCluster = eksCluster
	}
}

func WithEKSRoleARN(eksRoleARN string) Option {
	return func(k *kubeConfig) {
		k.EKSRoleARN = eksRoleARN
	}
}

func WithCertificate(certificate string) Option {
	return func(k *kubeConfig) {
		k.Certificate = certificate
	}
}

func WithSkipTLS(skipTLS bool) Option {
	return func(k *kubeConfig) {
		k.SkipTLS = skipTLS
	}
}

func WithNamespace(namespace string) Option {
	return func(k *kubeConfig) {
		k.Namespace = namespace
	}
}

func CreateKubeConfig(options ...Option) error {
	k := &kubeConfig{}
	for _, option := range options {
		option(k)
	}
	if k.Config == "" {
		return fmt.Errorf("no path to kubeconfig provided")
	}
	if k.ApiServer == "" {
		return fmt.Errorf("no kubernetes api server provided")
	}
	if k.Token == ""  && k.EKSCluster == "" {
		return fmt.Errorf("no kubernetes token provided")
	}
	if k.Token != "" && k.EKSCluster != "" {
		return fmt.Errorf("token cannot be used simultaneously with eksCluster")
	}
	if k.Namespace == "" {
		return fmt.Errorf("no namespace provided")
	}

	file, err := os.OpenFile(k.Config, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to write kubeconfig: %s", err)
	}
	defer file.Close()
	err = tmpl.Lookup("kubeconfig").Execute(file, k)
	if err != nil {
		return fmt.Errorf("unable to render kubeconfig: %s", err)
	}
	return nil
}

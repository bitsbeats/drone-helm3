package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/bitsbeats/drone-helm3/internal/helm"
	"github.com/bitsbeats/drone-helm3/internal/kube"
)

type (
	Config struct {
		KubeSkip        bool   `envconfig:"KUBE_SKIP" default:"false"`                // skip creation of kubeconfig
		KubeConfig      string `envconfig:"KUBE_CONFIG" default:"/root/.kube/config"` // path to kubeconfig
		KubeApiServer   string `envconfig:"KUBE_API_SERVER" required:"true"`          // kubernetes api server
		KubeToken       string `envconfig:"KUBE_TOKEN" required:"true"`               // kubernetes token
		KubeCertificate string `envconfig:"KUBE_CERTIFICATE"`                         // kubernetes http ca
		KubeSkipTLS     bool   `envconfig:"KUBE_SKIP_TLS" default:"false"`            // disable kubernetes tls verify

		Mode      string `envconfig:"MODE" default:"installupgrade"` // changes helm operation mode
		Chart     string `envconfig:"CHART" required:"true"`         // the helm chart to be deployed
		Release   string `envconfig:"RELEASE" required:"true"`       // helm release name
		Namespace string `envconfig:"NAMESPACE" required:"true"`     // kubernets and helm namespace

		Lint    bool `envconfig:"LINT" default:"true"`             // helm lint option
		Atomic  bool `envconfig:"ATOMIC" default:"true"`           // helm atomic option
		Wait    bool `envconfig:"WAIT" default:"true"`             // helm wait option
		Force   bool `envconfig:"FORCE" default:"false"`           // helm force option
		Cleanup bool `envconfig:"CLEANUP_ON_FAIL" default:"false"` // helm cleanup option
		DryRun  bool `envconfig:"DRY_RUN" default:"false"`         // helm dryrun option

		HelmRepos          []string `envconfig:"HELM_REPOS"`                          // additonal helm repos
		UpdateDependencies bool     `envconfig:"UPDATE_DEPENDENCIES" default:"false"` // helm update dependencies option

		Values     []string `envconfig:"VALUES"`      // additional --set options
		ValuesYaml string   `envconfig:"VAULES_YAML"` // additonal values files

		Timeout time.Duration `envconfig:"TIMEOUT" default:"15m"`  // timeout for helm command
		Debug   bool          `envconfig:"DEBUG", default:"false"` // debug configuration
	}
)

func main() {
	// lookup env file if specified
	envFile, ok := os.LookupEnv("PLUGIN_ENV_FILE")
	if ok {
		log.Printf("loading envfile %q", envFile)
		err := godotenv.Load(envFile)
		if err != nil {
			log.Printf("unable to load environmnet from file: %s", err)
		}
	}

	// load config from env
	cfg := &Config{}
	err := envconfig.Process("PLUGIN", cfg)
	if err != nil {
		log.Fatalf("unable to parse environment: %s", err)
	}

	// debug
	if cfg.Debug {
		debugCfg := *cfg
		debugCfg.KubeToken = "***"
		for i, val := range debugCfg.Values {
			kv := strings.SplitN(val, "=", 2)
			debugCfg.Values[i] = fmt.Sprintf("%s=***", kv[0])
		}
		log.Printf("configuration: %+v", debugCfg)
	}

	// create kube config
	if !cfg.KubeSkip {
		err = kube.CreateKubeConfig(
			kube.WithConfig(cfg.KubeConfig),
			kube.WithApiServer(cfg.KubeApiServer),
			kube.WithToken(cfg.KubeToken),
			kube.WithNamespace(cfg.Namespace),
			kube.WithCertificate(cfg.KubeCertificate),
			kube.WithSkipTLS(cfg.KubeSkipTLS),
		)
		if err != nil {
			log.Fatalf("unable to create kubernetes config: %s", err)
		}
	}

	// configure helm operation mode
	var modeOption helm.HelmModeOption
	switch cfg.Mode {
	case "installupgrade":
		modeOption = helm.WithInstallUpgradeMode()
	default:
		log.Fatalf("mode %q is not known", cfg.Mode)
	}

	// create helm cmd
	cmd, err := helm.NewHelmCmd(
		modeOption,
		helm.WithChart(cfg.Chart),
		helm.WithRelease(cfg.Release),
		helm.WithNamespace(cfg.Namespace),

		helm.WithLint(cfg.Lint),
		helm.WithAtomic(cfg.Atomic),
		helm.WithWait(cfg.Wait),
		helm.WithForce(cfg.Force),
		helm.WithCleanupOnFail(cfg.Cleanup),
		helm.WithDryRun(cfg.DryRun),

		helm.WithHelmRepos(cfg.HelmRepos),
		helm.WithUpdateDependencies(cfg.UpdateDependencies, cfg.Chart),

		helm.WithValues(cfg.Values),
		helm.WithValuesYaml(cfg.ValuesYaml),

		helm.WithKubeConfig(cfg.KubeConfig),
		helm.WithRunner(runner),
	)
	if err != nil {
		log.Fatalf("unable to generate helm command: %s", err)
	}

	// run commands
	log.Printf("running with a timeout of %s", cfg.Timeout.String())
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	err = cmd.Run(ctx)
	if err != nil {
		log.Fatalf("error running helm: %s", err)
	}
}

func runner(ctx context.Context, name string, args ...string) error {
	printArgs := make([]string, len(args))
	copy(printArgs, args)
	for i := 1; i < len(printArgs); i++ {
		if printArgs[i-1] == "--set-string" {
			kv := strings.SplitN(printArgs[i], "=", 2)
			printArgs[i] = fmt.Sprintf("%s=***", kv[0])
		}
	}
	log.Printf("running: %s %v", name, printArgs)

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	defer os.Stdout.Sync()
	defer os.Stderr.Sync()
	return cmd.Run()
}

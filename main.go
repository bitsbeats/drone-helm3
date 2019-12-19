package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/bitsbeats/drone-helm3/internal/helm"
	"github.com/bitsbeats/drone-helm3/internal/kube"
)

type (
	Config struct {
		KubeSkip        bool   `envconfig:"KUBE_SKIP" default:"false"`
		KubeConfig      string `envconfig:"KUBE_CONFIG" default:"/root/.kube/config"`
		KubeApiServer   string `envconfig:"KUBE_API_SERVER" required:"true"`
		KubeToken       string `envconfig:"KUBE_TOKEN" required:"true"`
		KubeCertificate string `envconfig:"KUBE_CERTIFICATE"`
		KubeSkipTLS     bool   `envconfig:"KUBE_SKIP_TLS" default:"false"`

		Mode      string `envconfig:"MODE" default:"installupgrade"`
		Chart     string `envconfig:"CHART" required:"true"`
		Release   string `envconfig:"RELEASE" required:"true"`
		Namespace string `envconfig:"NAMESPACE" required:"true"`

		Lint   bool `envconfig:"LINT" default:"true"`
		Wait   bool `envconfig:"WAIT" default:"true"`
		Force  bool `envconfig:"FORCE" default:"false"`
		DryRun bool `envconfig:"DRY_RUN" default:"false"`

		HelmRepos          map[string]string `envconfig:"HELM_REPOS"`
		UpdateDependencies bool              `envconfig:"UPDATE_DEPENDENCIES" default:"false"`

		Values     map[string]string `envconfig:"VALUES"`
		ValuesYaml string            `envconfig:"VAULES_YAML"`

		Timeout time.Duration `envconfig:"TIMEOUT" default:"15m"`
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

	// create kube config
	err = kube.CreateKubeConfig(
		kube.WithConfig(cfg.KubeConfig),
		kube.WithApiServer(cfg.KubeApiServer),
		kube.WithToken(cfg.KubeToken),
		kube.WithCertificate(cfg.KubeCertificate),
		kube.WithSkipTLS(cfg.KubeSkipTLS),
	)
	if err != nil {
		log.Fatalf("unable to create kubernetes config: %s", err)
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
		helm.WithWait(cfg.Wait),
		helm.WithForce(cfg.Force),
		helm.WithDryRun(cfg.DryRun),

		helm.WithHelmRepos(cfg.HelmRepos),
		helm.WithUpdateDependencies(cfg.UpdateDependencies),

		helm.WithValues(cfg.Values),
		helm.WithValuesYaml(cfg.ValuesYaml),

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
}

func runner(ctx context.Context, name string, args ...string) error {
	printArgs := make([]string, len(args))
	for i := 1; i < len(args); i++ {
		if args[i-1] == "--set-string" {
			kv := strings.SplitN(args[i], "=", 2)
			printArgs[i] = fmt.Sprintf("%s=***", kv[0])
		}
	}
	log.Printf("running: %s %v", name, printArgs)

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

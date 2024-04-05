package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/drone/envsubst"
	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/bitsbeats/drone-helm3/internal/errorhandler"
	"github.com/bitsbeats/drone-helm3/internal/helm"
	"github.com/bitsbeats/drone-helm3/internal/kube"
)

type (
	Config struct {
		PreCommands     string `envconfig:"PRE_COMMANDS" default:""`                  // can be used to run custom code, for example gcloud auth
		KubeSkip        bool   `envconfig:"KUBE_SKIP" default:"false"`                // skip creation of kubeconfig
		KubeConfig      string `envconfig:"KUBE_CONFIG" default:"/root/.kube/config"` // path to kubeconfig
		KubeApiServer   string `envconfig:"KUBE_API_SERVER"`                          // kubernetes api server
		KubeToken       string `envconfig:"KUBE_TOKEN"`                               // kubernetes token
		KubeCertificate string `envconfig:"KUBE_CERTIFICATE"`                         // kubernetes http ca
		KubeSkipTLS     bool   `envconfig:"KUBE_SKIP_TLS" default:"false"`            // disable kubernetes tls verify

		PushGatewayURL string `envconfig:"PUSHGATEWAY_URL" default:""` // url to a prometheus pushgateway server

		Mode      string `envconfig:"MODE" default:"installupgrade"` // changes helm operation mode
		Chart     string `envconfig:"CHART"`                         // the helm chart to be deployed
		Release   string `envconfig:"RELEASE" required:"true"`       // helm release name
		Namespace string `envconfig:"NAMESPACE" required:"true"`     // kubernets and helm namespace

		Lint                     bool   `envconfig:"LINT" default:"true"`                        // helm lint option
		Atomic                   bool   `envconfig:"ATOMIC" default:"true"`                      // helm atomic option
		Wait                     bool   `envconfig:"WAIT" default:"true"`                        // helm wait option
		Force                    bool   `envconfig:"FORCE" default:"false"`                      // helm force option
		Cleanup                  bool   `envconfig:"CLEANUP_ON_FAIL" default:"false"`            // helm cleanup option
		DryRun                   bool   `envconfig:"DRY_RUN" default:"false"`                    // helm dryrun option
		HelmDebug                bool   `envconfig:"HELM_DEBUG" default:"true"`                  // helm debug option
		DisableOpenAPIValidation bool   `envconfig:"DISABLE_OPENAPI_VALIDATION" default:"false"` // helm openapivalidation option
		PostKustomization        string `envconfig:"POST_KUSTOMIZATION" default:""`              // runs a customization of the generated output

		HelmRepos          []string `envconfig:"HELM_REPOS"`                          // additonal helm repos
		BuildDependencies  bool     `envconfig:"BUILD_DEPENDENCIES" default:"true"`   // helm dependency build option
		UpdateDependencies bool     `envconfig:"UPDATE_DEPENDENCIES" default:"false"` // helm dependency update option
		Test               bool     `envconfig:"TEST" default:"false"`                // helm run tests
		TestRollback       bool     `envconfig:"TEST_ROLLBACK" default:"false"`       // helm run tests and rollback on failure

		Envsubst             bool     `envconfig:"ENVSUBST" default:"false"`                // allow envsubst on Values und ValuesString
		Values               []string `envconfig:"VALUES"`                                  // additional --set options
		ValuesString         []string `envconfig:"VALUES_STRING"`                           // additional --set-string options
		ValuesYaml           string   `envconfig:"VALUES_YAML"`                             // additonal values files
		ValuesYamlAddDefault bool     `envconfig:"VALUES_YAML_ADD_DEFAULT" default:"false"` // re add the default values.yaml as first option

		Timeout time.Duration `envconfig:"TIMEOUT" default:"15m"` // timeout for helm command
		Debug   bool          `envconfig:"DEBUG" default:"false"` // debug configuration

		// auto-filled by drone
		DroneRepo string `envconfig:"DRONE_REPO" required:"true"`
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

	var eh errorhandler.Handler
	if cfg.PushGatewayURL != "" {
		log.Printf("pushgateway is %s", cfg.PushGatewayURL)
		eh = errorhandler.NewPushgateway(cfg.DroneRepo, cfg.Namespace, cfg.Release, cfg.PushGatewayURL)
	} else {
		eh = errorhandler.NewLog()
	}

	// debug
	if cfg.Debug {
		debugCfg := Config{}
		_ = copier.Copy(&debugCfg, cfg)
		debugCfg.KubeToken = "***"
		for i, val := range debugCfg.Values {
			kv := strings.SplitN(val, "=", 2)
			debugCfg.Values[i] = fmt.Sprintf("%s=***", kv[0])
		}
		for i, val := range debugCfg.ValuesString {
			kv := strings.SplitN(val, "=", 2)
			debugCfg.Values[i] = fmt.Sprintf("%s=***", kv[0])
		}
		log.Printf("configuration: %+v", debugCfg)
	}

	// run pre commands if set
	if cfg.PreCommands != "" {
		scriptName := "/tmp/pre_commands.sh"
		f, err := os.Create(scriptName)
		if err != nil {
			log.Fatalf("unable to create precommands file: %s", err)
		}
		_, err = f.WriteString(cfg.PreCommands)
		if err != nil {
			log.Fatalf("unable to write precommands to file: %s", err)
		}

		cmd := exec.Command("/bin/bash", scriptName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("unable to run pre commands: %s", err)
		}
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
			eh.Fatalf("unable to create kubernetes config: %s", err)
		}
	}

	// envsubst
	if cfg.Envsubst {
		log.Print("envsubst is enabled")
		var err error
		for i, val := range cfg.Values {
			cfg.Values[i], err = envsubst.EvalEnv(val)
			if err != nil {
				eh.Fatalf("unable to envsubst %s: %s", val, err)
			}
		}
		for i, val := range cfg.ValuesString {
			cfg.ValuesString[i], err = envsubst.EvalEnv(val)
			if err != nil {
				eh.Fatalf("unable to envsubst %s: %s", val, err)
			}
		}
	}

	// configure helm operation mode
	var cmd *helm.HelmCmd
	switch cfg.Mode {
	case "installupgrade":
		// helm validations
		// no need to download old versions if we update
		if cfg.UpdateDependencies {
			cfg.BuildDependencies = false
		}
		// test rollback requires test
		if cfg.TestRollback {
			cfg.Test = true
		}

		// create helm cmd
		cmd, err = helm.NewHelmCmd(
			helm.WithInstallUpgradeMode(),
			helm.WithChart(cfg.Chart),
			helm.WithRelease(cfg.Release),
			helm.WithNamespace(cfg.Namespace),

			helm.WithTimeout(cfg.Timeout),
			helm.WithAtomic(cfg.Atomic),
			helm.WithWait(cfg.Wait),
			helm.WithForce(cfg.Force),
			helm.WithCleanupOnFail(cfg.Cleanup),
			helm.WithDryRun(cfg.DryRun),
			helm.WithDebug(cfg.HelmDebug),
			helm.WithDisableOpenAPIValidation(cfg.DisableOpenAPIValidation),
			helm.WithPostKustomization(cfg.PostKustomization),

			helm.WithHelmRepos(cfg.HelmRepos),
			helm.WithBuildDependencies(cfg.BuildDependencies, cfg.Chart),
			helm.WithUpdateDependencies(cfg.UpdateDependencies, cfg.Chart),
			helm.WithLint(cfg.Lint),
			helm.WithTest(cfg.Test, cfg.Release),
			helm.WithTestRollback(cfg.Test, cfg.Release),

			helm.WithValuesYamlAddDefault(cfg.ValuesYamlAddDefault, cfg.Chart),
			helm.WithValuesYaml(cfg.ValuesYaml),
			helm.WithValues(cfg.Values),
			helm.WithValuesString(cfg.ValuesString),

			helm.WithKubeConfig(cfg.KubeConfig),
			helm.WithRunner(NewRunner()),
		)
		if err != nil {
			eh.Fatalf("unable to generate helm command: %s", err)
		}
	case "uninstall":
		cmd, err = helm.NewHelmCmd(
			helm.WithUninstallMode(),
			helm.WithRelease(cfg.Release),
			helm.WithNamespace(cfg.Namespace),

			helm.WithWait(cfg.Wait),
			helm.WithTimeout(cfg.Timeout),

			helm.WithKubeConfig(cfg.KubeConfig),
			helm.WithRunner(NewRunner()),
		)
		if err != nil {
			eh.Fatalf("unable to generate helm command: %s", err)
		}
	default:
		eh.Fatalf("mode %q is not known", cfg.Mode)
	}

	// run commands
	log.Printf("running with a timeout of %s", cfg.Timeout.String())
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout+(10*time.Minute))
	defer cancel()
	err = cmd.Run(ctx)
	if err != nil {
		eh.Fatalf("error running helm: %s", err)
	}
	eh.Status(err, "finished deployment successfully")
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, name string, args ...string) error {
	printArgs := make([]string, len(args))
	copy(printArgs, args)
	for i := 1; i < len(printArgs); i++ {
		if printArgs[i-1] == "--set-string" || printArgs[i-1] == "--set" {
			kv := strings.SplitN(printArgs[i], "=", 2)
			printArgs[i] = fmt.Sprintf("%s=***", kv[0])
		}
	}
	log.Printf("running: %s %v", name, printArgs)

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stderr = os.Stderr
	defer os.Stdout.Sync()
	defer os.Stderr.Sync()
	return cmd.Run()
}

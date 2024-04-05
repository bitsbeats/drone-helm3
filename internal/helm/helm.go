package helm

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bitsbeats/drone-helm3/internal/core"
)

type (
	HelmCmd struct {
		Mode HelmMode

		Release string
		Chart   string
		Args    []string

		PreCmds  [][]string
		PostCmds [][]string
		Runner   Runner

		Test         bool
		TestRollback bool

		OnSuccess                   []func()
		OnTestSuccess               []func()
		OnTestFailed                []func()
		OnTestFailedRollbackSuccess []func()
		OnTestFailedRollbackFailed  []func()
	}

	HelmMode = string

	// Build pattern options
	HelmModeOption func(*HelmCmd)
	HelmOption     func(*HelmCmd) error
	Runner         interface {
		Run(ctx context.Context, command string, args ...string) error
	}
)

const (
	InstallUpgradeMode HelmMode = "install-upgrade"
	UninstallMode      HelmMode = "uninstall"
)

func WithInstallUpgradeMode() HelmModeOption {
	return func(c *HelmCmd) {
		c.Mode = InstallUpgradeMode
		c.Args = append([]string{"upgrade", "--install"}, c.Args...)
	}
}

func WithUninstallMode() HelmModeOption {
	return func(c *HelmCmd) {
		c.Mode = UninstallMode
		c.Args = append([]string{"uninstall"}, c.Args...)
	}
}

func WithRelease(release string) HelmOption {
	return func(c *HelmCmd) error {
		c.Release = release
		return nil
	}
}

func WithChart(chart string) HelmOption {
	return func(c *HelmCmd) error {
		c.Chart = chart
		return nil
	}
}

func WithNamespace(namespace string) HelmOption {
	return func(c *HelmCmd) error {
		c.Args = append(c.Args, "-n", namespace)
		return nil
	}
}

func WithLint(lint bool) HelmOption {
	return func(c *HelmCmd) error {
		if lint {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "lint", c.Chart,
			})
		}
		return nil
	}
}

func WithAtomic(atomic bool) HelmOption {
	return func(c *HelmCmd) error {
		if atomic {
			c.Args = append(c.Args, "--atomic")
		}
		return nil
	}
}

func WithWait(wait bool) HelmOption {
	return func(c *HelmCmd) error {
		if wait {
			c.Args = append(c.Args, "--wait")
		}
		return nil
	}
}

func WithForce(force bool) HelmOption {
	return func(c *HelmCmd) error {
		if force {
			c.Args = append(c.Args, "--force")
		}
		return nil
	}
}

func WithCleanupOnFail(cleanup bool) HelmOption {
	return func(c *HelmCmd) error {
		if cleanup {
			c.Args = append(c.Args, "--cleanup-on-fail")
		}
		return nil
	}
}

func WithDryRun(dry bool) HelmOption {
	return func(c *HelmCmd) error {
		if dry {
			c.Args = append(c.Args, "--dry-run")
		}
		return nil
	}
}

func WithDebug(dry bool) HelmOption {
	return func(c *HelmCmd) error {
		if dry {
			c.Args = append(c.Args, "--debug")
		}
		return nil
	}
}

func WithDisableOpenAPIValidation(disable bool) HelmOption {
	return func(c *HelmCmd) error {
		if disable {
			c.Args = append(c.Args, "--disable-openapi-validation")
		}
		return nil
	}
}

func WithPostKustomization(kustomization string) HelmOption {
	return func(c *HelmCmd) error {
		if kustomization != "" {
			f, err := os.OpenFile("/kustomize/kustomization.yaml", os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				return fmt.Errorf("unable to create kustomization file: %w", err)
			}
			_, err = f.WriteString(kustomization)
			if err != nil {
				return fmt.Errorf("unable to write to kustomization file: %w", err)
			}
			c.Args = append(c.Args, "--post-renderer", "/kustomize/kustomize.sh")
		}
		return nil
	}
}

func WithTimeout(timeout time.Duration) HelmOption {
	return func(c *HelmCmd) error {
		c.Args = append(c.Args, "--timeout", timeout.String())
		return nil
	}
}

func WithHelmRepos(repos []string) HelmOption {
	return func(c *HelmCmd) error {
		if len(repos) == 0 {
			return nil
		}
		for _, repo := range repos {
			split := strings.SplitN(repo, "=", 2)
			if len(split) != 2 {
				return fmt.Errorf("not in key=value format: %s", repo)
			}
			name := split[0]
			url := split[1]
			log.Printf("added repo: name:%q url:%q", name, url)
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "repo", "add", name, url,
			})
		}
		c.PreCmds = append(c.PreCmds, []string{
			"helm", "repo", "update",
		})
		return nil
	}
}

func WithBuildDependencies(build bool, chart string) HelmOption {
	return func(c *HelmCmd) error {
		if build {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "dependency", "build", chart,
			})
		}
		return nil
	}
}

func WithUpdateDependencies(update bool, chart string) HelmOption {
	return func(c *HelmCmd) error {
		if update {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "dependency", "update", chart,
			})
		}
		return nil
	}
}

func WithTest(test bool, release string) HelmOption {
	return func(c *HelmCmd) error {
		c.Test = test
		return nil
	}
}

func WithTestRollback(test bool, release string) HelmOption {
	return func(c *HelmCmd) error {
		c.TestRollback = test
		return nil
	}
}

func WithValues(values []string) HelmOption {
	return func(c *HelmCmd) error {
		for _, v := range values {
			split := strings.SplitN(v, "=", 2)
			if len(split) != 2 {
				return fmt.Errorf("not in key=value format: %s", v)
			}
			key := split[0]
			value := split[1]
			c.Args = append(c.Args, "--set", fmt.Sprintf("%s=%s", key, value))
		}
		return nil
	}
}

func WithValuesString(values []string) HelmOption {
	return func(c *HelmCmd) error {
		for _, v := range values {
			split := strings.SplitN(v, "=", 2)
			if len(split) != 2 {
				return fmt.Errorf("not in key=value format: %s", v)
			}
			key := split[0]
			value := split[1]
			c.Args = append(c.Args, "--set-string", fmt.Sprintf("%s=%s", key, value))
		}
		return nil
	}
}

func WithValuesYaml(file string) HelmOption {
	return func(c *HelmCmd) error {
		if file != "" {
			c.Args = append(c.Args, "--values", file)
		}
		return nil
	}
}

func WithValuesYamlAddDefault(add bool, chartpath string) HelmOption {
	return func(c *HelmCmd) error {
		if add {
			file := fmt.Sprintf("%s/values.yaml", chartpath)
			_, err := os.Stat(file)
			if os.IsNotExist(err) {
				return fmt.Errorf("unable to find Default values file: %s", err)
			}
			c.Args = append(c.Args, "--values", file)
		}
		return nil
	}
}

func WithPreCommand(command ...string) HelmOption {
	return func(c *HelmCmd) error {
		c.PreCmds = append(c.PreCmds, command)
		return nil
	}
}

func WithPostCommand(command ...string) HelmOption {
	return func(c *HelmCmd) error {
		c.PostCmds = append(c.PostCmds, command)
		return nil
	}
}

func WithKubeConfig(config string) HelmOption {
	return func(c *HelmCmd) error {
		if config != "" {
			c.Args = append(c.Args, "--kubeconfig", config)
		}
		return nil
	}
}

func WithRunner(runner Runner) HelmOption {
	return func(c *HelmCmd) error {
		c.Runner = runner
		return nil
	}
}

func NewHelmCmd(mode HelmModeOption, options ...HelmOption) (*HelmCmd, error) {
	h := &HelmCmd{
		Args:     []string{},
		PreCmds:  [][]string{},
		PostCmds: [][]string{},
		Runner:   nil,
	}
	mode(h)
	for _, option := range options {
		err := option(h)
		if err != nil {
			return nil, fmt.Errorf("unable to parse option: %s", err)
		}
	}
	if h.Release == "" {
		return nil, fmt.Errorf("release name is required")
	}
	if h.Chart == "" && h.Mode != UninstallMode {
		return nil, fmt.Errorf("chart path is required")
	}
	if h.Runner == nil {
		return nil, fmt.Errorf("runner is required")
	}

	switch h.Mode {
	case InstallUpgradeMode:
		h.Args = append(h.Args, h.Release, h.Chart)
	case UninstallMode:
		h.Args = append(h.Args, h.Release)
	default:
		return nil, fmt.Errorf("mode %q is not known", h.Mode)
	}
	return h, nil
}

func (h *HelmCmd) Run(ctx context.Context) error {
	for _, preCmd := range h.PreCmds {
		err := h.Runner.Run(ctx, preCmd[0], preCmd[1:]...)
		if err != nil {
			return Wrap(err, "precmd failed", core.PreFailErrorKind)
		}
	}
	err := h.Runner.Run(ctx, "helm", h.Args...)
	if err != nil {
		return Wrap(err, "helm failed", core.FailedErrorKind)
	}
	if h.Test {
		err := h.Runner.Run(ctx, "helm", "test", "--logs", h.Release)
		if err != nil {
			log.Printf("TEST FAILED: %s", err)
			if h.TestRollback {
				rollbackErr := h.Runner.Run(ctx, "helm", "rollback", h.Release)
				if rollbackErr != nil {
					log.Printf("ROLLBACK FAILED: %s", rollbackErr)
					return Wrap(rollbackErr, "release and rollback failed", core.RollbackFailedErrorKind)
				} else {
					log.Printf("TEST FAILED: %s", err)
				}
			}
			return Wrap(err, "release failed and rollback successful", core.RollbackSuccessErrorKind)
		}
	}
	for _, postCmd := range h.PostCmds {
		err := h.Runner.Run(ctx, postCmd[0], postCmd[1:]...)
		if err != nil {
			return Wrap(err, "postcmd failed", core.PostFailErrorKind)
		}
	}
	return nil
}

type (
	HelmError struct {
		Context string
		Kind    core.ErrorKind
		Err     error
	}
)

func (e *HelmError) Error() string {
	return fmt.Sprintf("%s: %s", e.Context, e.Err)
}

func Wrap(err error, info string, kind core.ErrorKind) *HelmError {
	return &HelmError{
		Context: info,
		Kind:    kind,
		Err:     err,
	}
}

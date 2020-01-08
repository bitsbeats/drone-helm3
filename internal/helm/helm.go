package helm

import (
	"context"
	"fmt"
	"log"
	"strings"
)

type (
	HelmCmd struct {
		Release string
		Chart   string
		Args    []string

		PreCmds  [][]string
		PostCmds [][]string
		Runner   Runner
	}

	HelmOption     func(*HelmCmd) error
	HelmModeOption func(*HelmCmd)
	Runner         func(ctx context.Context, command string, args ...string) error
)

func WithInstallUpgradeMode() HelmModeOption {
	return func(c *HelmCmd) {
		c.Args = append([]string{"upgrade", "--install"}, c.Args...)
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

func WithHelmRepos(repos []string) HelmOption {
	return func(c *HelmCmd) error {
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

func WithValues(values []string) HelmOption {
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
	if h.Chart == "" {
		return nil, fmt.Errorf("chart path is required")
	}
	if h.Runner == nil {
		return nil, fmt.Errorf("runner is required")
	}
	h.Args = append(h.Args, h.Release, h.Chart)
	return h, nil
}

func (h *HelmCmd) Run(ctx context.Context) error {
	for _, preCmd := range h.PreCmds {
		err := h.Runner(ctx, preCmd[0], preCmd[1:]...)
		if err != nil {
			return fmt.Errorf("precmd failed: %s", err)
		}
	}
	err := h.Runner(ctx, "helm", h.Args...)
	if err != nil {
		return fmt.Errorf("helm failed: %s", err)
	}
	for _, postCmd := range h.PostCmds {
		err := h.Runner(ctx, postCmd[0], postCmd[1:]...)
		if err != nil {
			return fmt.Errorf("postcmd failed: %s", err)
		}
	}
	return nil
}

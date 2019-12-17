package helm

import (
	"context"
	"fmt"
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

	HelmOption func(*HelmCmd)
	Runner     func(ctx context.Context, command string, args ...string) error
)

func WithRelease(release string) HelmOption {
	return func(c *HelmCmd) {
		c.Release = release
	}
}

func WithChart(chart string) HelmOption {
	return func(c *HelmCmd) {
		c.Chart = chart
	}
}

func WithNamespace(namespace string) HelmOption {
	return func(c *HelmCmd) {
		c.Args = append(c.Args, "-n", namespace)
	}
}

func WithLint(lint bool) HelmOption {
	return func(c *HelmCmd) {
		if lint {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "lint",
			})
		}
	}
}

func WithWait(wait bool) HelmOption {
	return func(c *HelmCmd) {
		if wait {
			c.Args = append(c.Args, "--wait")
		}
	}
}

func WithForce(force bool) HelmOption {
	return func(c *HelmCmd) {
		if force {
			c.Args = append(c.Args, "--force")
		}
	}
}

func WithDryRun(dry bool) HelmOption {
	return func(c *HelmCmd) {
		if dry {
			c.Args = append(c.Args, "--dry-run")
		}
	}
}

func WithHelmRepos(repos map[string]string) HelmOption {
	return func(c *HelmCmd) {
		for name, url := range repos {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "repo", "add", name, url,
			})
		}
		c.PreCmds = append(c.PreCmds, []string{
			"helm", "repo", "update",
		})
	}
}

func WithUpdateDependencies(update bool) HelmOption {
	return func(c *HelmCmd) {
		if update {
			c.PreCmds = append(c.PreCmds, []string{
				"helm", "dependeny", "update",
			})
		}
	}
}

func WithValues(values map[string]string) HelmOption {
	return func(c *HelmCmd) {
		for key, value := range values {
			c.Args = append(c.Args, "--set-string", fmt.Sprintf("%s=%s", key, value))
		}
	}
}

func WithValuesYaml(file string) HelmOption {
	return func(c *HelmCmd) {
		c.Args = append(c.Args, "--values", file)
	}
}

func WithRunner(runner Runner) HelmOption {
	return func(c *HelmCmd) {
		c.Runner = runner
	}
}

func NewHelmCmd(options ...HelmOption) (*HelmCmd, error) {
	h := &HelmCmd{}
	for _, option := range options {
		option(h)
	}
	if h.Release == "" {
		return nil, fmt.Errorf("release name is required")
	}
	if h.Chart == "" {
		return nil, fmt.Errorf("chart path is required")
	}
	if h.Runner == nil {
		return nil, fmt.Errorf("unable to run without runner")
	}
	return h, nil
}

func (h *HelmCmd) Run(ctx context.Context) error {
	for _, preCmd := range h.PreCmds {
		err := h.Runner(ctx, preCmd[0], preCmd[1:]...)
		if err != nil {
			return err
		}
	}
	err := h.Runner(ctx, "helm", h.Args...)
	if err != nil {
		return err
	}
	for _, postCmd := range h.PostCmds {
		err := h.Runner(ctx, postCmd[0], postCmd[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}

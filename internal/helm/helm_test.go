package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestHelmCmd(t *testing.T) {
	var w io.Writer
	runner := func(ctx context.Context, cmd string, args ...string) error {
		fmt.Fprintf(w, "%s %s\n", cmd, strings.Join(args, " "))
		return nil
	}
	tests := []struct {
		mode      HelmModeOption
		options   []HelmOption
		want      string
		createErr error
		runErr    error
	}{
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithRunner(runner),
			},
			want: "helm upgrade --install foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithLint(true),
				WithRunner(runner),
			},
			want: "helm lint chart\n" +
				"helm upgrade --install foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithDryRun(true),
				WithRunner(runner),
			},
			want: "helm upgrade --install --dry-run foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithBuildDependencies(true, "chart"),
				WithRunner(runner),
			},
			want: "helm dependency build chart\n" +
				"helm upgrade --install foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithUpdateDependencies(true, "chart"),
				WithRunner(runner),
			},
			want: "helm dependency update chart\n" +
				"helm upgrade --install foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithHelmRepos([]string{
					"dev=https://example.com/dev-charts",
				}),
				WithRunner(runner),
			},
			want: "helm repo add dev https://example.com/dev-charts\n" +
				"helm repo update\n" +
				"helm upgrade --install foo chart\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-staging"),
				WithRelease("myapp-staging"),
				WithChart("./helm/myapp"),
				WithWait(true),
				WithForce(true),
				WithValuesYaml("./helm/values.yaml"),
				WithValues([]string{
					"git.commit_sha=21ffea3",
				}),
				WithRunner(runner),
			},
			want: "helm upgrade --install -n myapp-staging --wait --force --values ./helm/values.yaml " +
				"--set git.commit_sha=21ffea3 myapp-staging ./helm/myapp\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithWait(true),
				WithValuesYaml("./helm/values.yaml"),
				WithValuesString([]string{
					"git.commit_sha=21efea3",
				}),
				WithRunner(runner),
			},
			want: "helm upgrade --install -n myapp-production --wait --values ./helm/values.yaml " +
				"--set-string git.commit_sha=21efea3 myapp-production ./helm/myapp\n",
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("myapp"),
				WithChart("./helm/myapp"),
			},
			createErr: fmt.Errorf("runner is required"),
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("myapp"),
				WithRunner(runner),
			},
			createErr: fmt.Errorf("chart path is required"),
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithChart("./helm/myapp"),
				WithRunner(runner),
			},
			createErr: fmt.Errorf("release name is required"),
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithPreCommand("prefail"),
				WithRunner(func(ctx context.Context, name string, args ...string) error {
					if name == "prefail" {
						return fmt.Errorf("prefail")
					}
					return nil
				}),
			},
			runErr: fmt.Errorf("precmd failed: prefail"),
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithRunner(func(ctx context.Context, name string, args ...string) error {
					if name == "helm" {
						return fmt.Errorf("runfail")
					}
					return nil
				}),
			},
			runErr: fmt.Errorf("helm failed: runfail"),
		},
		{
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithPostCommand("postfail"),
				WithRunner(func(ctx context.Context, name string, args ...string) error {
					if name == "postfail" {
						return fmt.Errorf("postfail")
					}
					return nil
				}),
			},
			runErr: fmt.Errorf("postcmd failed: postfail"),
		},
	}

	for _, test := range tests {
		w = bytes.NewBuffer([]byte{})
		cmd, err := NewHelmCmd(test.mode, test.options...)
		if !errEq(err, test.createErr) {
			t.Fatalf("unable to create helm cmd:\n- %v\n+ %v", test.createErr, err)
		} else if err != nil {
			continue
		}
		err = cmd.Run(context.Background())
		if !errEq(err, test.runErr) {
			t.Fatalf("unable to run helm cmd:\n- %v\n+ %v", test.runErr, err)
		} else if err != nil {
			continue
		}
		got := w.(*bytes.Buffer).String()
		if test.want != got {
			t.Fatalf("mismatch:\n- %s\n+ %s", test.want, got)
		}
	}
}

func errEq(a error, b error) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

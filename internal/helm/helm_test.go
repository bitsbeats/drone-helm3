package helm

import (
	"context"
	"fmt"
	"testing"

	"github.com/bitsbeats/drone-helm3/mock"
	"github.com/golang/mock/gomock"
)

func TestHelmCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRunner := mock.NewMockRunner(ctrl)

	tests := []struct {
		name      string
		mode      HelmModeOption
		options   []HelmOption
		setup     func()
		createErr error
		runErr    error
	}{
		{
			name: "helm upgrade",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "foo", "chart",
				)
			},
		},
		{
			name: "helm upgrade with lint",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithLint(true),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "lint", "chart",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "foo", "chart",
				)

			},
		},
		{
			name: "helm upgrade as dry run",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithDryRun(true),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "--dry-run", "foo", "chart",
				)
			},
		},
		{
			name: "helm upgrade with dependency build",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithBuildDependencies(true, "chart"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "dependency", "build", "chart",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "foo", "chart",
				)
			},
		},
		{
			name: "helm upgrade with dependency update",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithUpdateDependencies(true, "chart"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "dependency", "update", "chart",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "foo", "chart",
				)
			},
		},
		{
			name: "helm upgrade with repos",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("foo"),
				WithChart("chart"),
				WithHelmRepos([]string{
					"dev=https://example.com/dev-charts",
				}),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "repo", "add", "dev", "https://example.com/dev-charts",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "repo", "update",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "foo", "chart",
				)
			},
		},
		{
			name: "helm upgrade with force, wait, vaules and values yaml",
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
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-staging",
					"--wait", "--force", "--values", "./helm/values.yaml",
					"--set", "git.commit_sha=21ffea3",
					"myapp-staging", "./helm/myapp",
				)
			},
		},
		{
			name: "helm upgrade with wait, values yaml and values string",
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
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"--wait", "--values", "./helm/values.yaml",
					"--set-string", "git.commit_sha=21efea3",
					"myapp-production", "./helm/myapp",
				)
			},
		},
		{
			name: "with missing runner",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("myapp"),
				WithChart("./helm/myapp"),
			},
			createErr: fmt.Errorf("runner is required"),
		},
		{
			name: "with missing chart",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithRelease("myapp"),
				WithRunner(mockRunner),
			},
			createErr: fmt.Errorf("chart path is required"),
		},
		{
			name: "with missing release name",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithChart("./helm/myapp"),
				WithRunner(mockRunner),
			},
			createErr: fmt.Errorf("release name is required"),
		},
		{
			name: "with failed precmd",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithPreCommand("prefail"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"prefail",
				).Return(fmt.Errorf("prefail"))
			},
			runErr: fmt.Errorf("precmd failed: prefail"),
		},
		{
			name: "with failed helm upgrade",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				).Return(fmt.Errorf("runfail"))
			},
			runErr: fmt.Errorf("helm failed: runfail"),
		},
		{
			name: "with failed postcmd",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithPostCommand("postfail"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"postfail",
				).Return(fmt.Errorf("postfail"))
			},
			runErr: fmt.Errorf("postcmd failed: postfail"),
		},
		{
			name: "with helm test",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithTest(true, "myapp-release"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "test", "--logs", "myapp-production",
				)
			},
			runErr: nil,
		},
		{
			name: "with failed helm test",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithTest(true, "myapp-release"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "test", "--logs", "myapp-production",
				).Return(fmt.Errorf("testfail"))
			},
			runErr: fmt.Errorf("release failed and rollback successful: testfail"),
		},
		{
			name: "with failed test and sucessfull rollback",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithTest(true, "myapp-release"),
				WithTestRollback(true, "myapp-release"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "test", "--logs", "myapp-production",
				).Return(fmt.Errorf("testfail"))
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "rollback", "myapp-production",
				)
			},
			runErr: fmt.Errorf("release failed and rollback successful: testfail"),
		},
		{
			name: "with failed test and failed rollback",
			mode: WithInstallUpgradeMode(),
			options: []HelmOption{
				WithNamespace("myapp-production"),
				WithRelease("myapp-production"),
				WithChart("./helm/myapp"),
				WithTest(true, "myapp-release"),
				WithTestRollback(true, "myapp-release"),
				WithRunner(mockRunner),
			},
			setup: func() {
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "upgrade", "--install", "-n", "myapp-production",
					"myapp-production", "./helm/myapp",
				)
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "test", "--logs", "myapp-production",
				).Return(fmt.Errorf("testfail"))
				mockRunner.EXPECT().Run(
					context.Background(),
					"helm", "rollback", "myapp-production",
				).Return(fmt.Errorf("rollbackfail"))
			},
			runErr: fmt.Errorf("release and rollback failed: rollbackfail"),
		},
	}

	for i, test := range tests {
		t.Logf("running #%d: %s", i, test.name)
		cmd, err := NewHelmCmd(test.mode, test.options...)
		if !errEq(err, test.createErr) {
			t.Fatalf("unable to create helm cmd:\n- %v\n+ %v", test.createErr, err)
		} else if err != nil {
			continue
		}
		test.setup()
		err = cmd.Run(context.Background())
		if !errEq(err, test.runErr) {
			t.Fatalf("unable to run helm cmd:\n- %v\n+ %v", test.runErr, err)
		}
	}
}

func errEq(a error, b error) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

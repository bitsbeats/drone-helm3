# drone-helm3

[![Build Status](https://cloud.drone.io/api/badges/bitsbeats/drone-helm3/status.svg)](https://cloud.drone.io/bitsbeats/drone-helm3)
[![Docker Pulls](https://img.shields.io/docker/pulls/bitsbeats/drone-helm3.svg?maxAge=604800)](https://hub.docker.com/r/bitsbeats/drone-helm3)
[![Go Report Card](https://goreportcard.com/badge/github.com/bitsbeats/drone-helm3)](https://goreportcard.com/report/github.com/bitsbeats/drone-helm3)

Drone plugin for Helm3.

Helm Version: 3.9.0  
Kubectl Version: 1.22.11

## Drone settings

Example:

```yaml
- name: deploy app
  image: bitsbeats/drone-helm3
  settings:
    kube_api_server: kube.example.com
    kube_token: { from_secret: kube_token }

    chart: ./path-to/chart
    release: release-name
    namespace: namespace-name
    timeout: 20m
    helm_repos:
      - bitnami=https://charts.bitnami.com/bitnami
    envsubst: true
    values:
      - app.environment=awesome
      - app.tag=${DRONE_TAG/v/}
      - app.commit=${DRONE_COMMIT_SHA}
```

**Note**: If you enable envsubst make sure to surrount your variables like
`${variable}`, `$variable` will *not* work.

An always up2date version of the availible config options can be viewed on the
source on the `Config` `struct` [here][1].

## Monitoring

Its possible to monitor your builds and rollbacks using prometheus and
prometheus-pushgateway. To enable specify the `pushgateway_url` setting.

Example alertrule:

```
          - alert: Helm3RolloutFailed
            expr: |
              drone_helm3_build_status{status!="success"}
            labels:
              severity: critical
            annotations:
              summary: >-
                Helm3 was unable to deploy {{ $labels.repo }} as
                {{ $labels.release }} into namespace {{ $labels.namespace }}
              action: >-
                Validate the `deploy` step of the last drone ci run for this
                repository. Either the build has *failed entirely* or the
                `helm test` did fail. For more information on tests see
                https://github.com/bitsbeats/drone-helm3/#monitoring
```

## Helm Tests

Helm tests are special Pods that have the `"helm.sh/hook": test` annotation set.
If the command in the docker container returns an exitcode > 0 the drone step
will be marked as failed. See the [Helm documentation][2].

In addition you can set the `test_rollback` setting to run `helm rollback` if
the tests fail.

## `post_kustomization`

The `post_kustomization` allows to modify helm charts with customize.
See [here][3] for the official documentation. The `resources` field is
set via the plugin.

Example:

```yaml
post_kustomization: |
  patches:
    - patch: |
        - op: remove
          path: /spec/template/spec/securityContext
        - op: remove
          path: /spec/template/spec/containers/0/securityContext
      target:
        kind: StatefulSet
        labelSelector:
          app.kubernetes.io/name=opensearch
```


[1]: https://github.com/bitsbeats/drone-helm3/blob/master/main.go#L22
[2]: https://helm.sh/docs/topics/chart_tests/
[3]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/

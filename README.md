# drone-helm3

[![Build Status](https://cloud.drone.io/api/badges/bitsbeats/drone-helm3/status.svg)](https://cloud.drone.io/bitsbeats/drone-helm3)
[![Docker Pulls](https://img.shields.io/docker/pulls/bitsbeats/drone-helm3.svg?maxAge=604800)](https://hub.docker.com/r/bitsbeats/drone-helm3)
[![Go Report Card](https://goreportcard.com/badge/github.com/bitsbeats/drone-helm3)](https://goreportcard.com/report/github.com/bitsbeats/drone-helm3)

Drone plugin for Helm3.

Helm Version: 3.3.4  
Kubectl Version: 1.19.2

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

Following settings are availible as Drones `settings:`, a full list can be
viewed on the `Config` `struct`
[here](https://github.com/bitsbeats/drone-helm3/blob/master/main.go#L22).

Kubernetes:

* `kube_skip`: skip creation of kubeconfig (**optional**, **default**:`false`)
* `kube_config`: path to kubeconfig (**optional**, **default**:`/root/.kube/config`)
* `kube_api_server`: kubernetes api server (**required**)
* `kube_token`: kubernetes token (**required**)
* `kube_certificate`: kubernetes http ca (**optional**)
* `kube_skip_tls`: disable kubernetes tls verify (**optional**, **default**:`false`)

Helm:

* `mode`: changes helm operation mode (**optional**, **default**:`installupgrade`)
* `chart`: the helm chart to be deployed (**required**)
* `release`: helm release name (**required**)
* `namespace`: kubernets and helm namespace (**required**)
* `lint`: helm lint option (**optional**, **default**:`true`)
* `atomic`: helm atomic option (**optional**, **default**:`true`)
* `wait`: helm wait option (**optional**, **default**:`true`)
* `force`: helm force option (**optional**, **default**:`false`)
* `cleanup_on_fail`: helm cleanup option (**optional**, **default**:`false`)
* `dry_run`: helm dryrun option (**optional**, **default**:`false`)
* `helm_debug`: helm debug option (**optional**, **default**:`true`)
* `helm_repos`: additonal helm repos (**optional**)
* `build_dependencies`: helm dependency build option (**optional**, **default**:`true`)
* `update_dependencies`: helm dependency update option (**optional**, **default**:`false`, **disables** `BUILD_DEPENDENCIES`)
* `test`: run helm tests after the helm upgrade (**optional**, **default**: `false`)
* `test_rollback`: run helm tests and rollback if the tests fail (**optional**, **default**: `false`)
* `envsubst`: allow envsubst on Values und ValuesString (**optional**, **default**:`false`)
* `values`: additional --set options (**optional**)
* `values_string`: additional --set-string options (**optional**)
* `values_yaml`: additonal values files (**optional**)

General:

* `timeout`: timeout for helm command (**optional**, **default**:`15m`)

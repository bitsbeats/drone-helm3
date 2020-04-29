# drone-helm3

[![Build Status](https://cloud.drone.io/api/badges/bitsbeats/drone-helm3/status.svg)](https://cloud.drone.io/bitsbeats/drone-helm3)
[![Docker Pulls](https://img.shields.io/docker/pulls/bitsbeats/drone-helm3.svg?maxAge=604800)](https://hub.docker.com/r/bitsbeats/drone-helm3)
[![Go Report Card](https://goreportcard.com/badge/github.com/bitsbeats/drone-helm3)](https://goreportcard.com/report/github.com/bitsbeats/drone-helm3)

Drone plugin for Helm3.

Helm Version: 3.1.3  
Kubectl Version: 1.17.2

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

**Note**: If you enable envsubst make sure to surrount your variables like `${variable}`, `$variable` will *not* work.

Kubernetes:

* `KUBE_SKIP`: skip creation of kubeconfig (**optional**, **default**:`false`)
* `KUBE_CONFIG`: path to kubeconfig (**optional**, **default**:`/root/.kube/config`)
* `KUBE_API_SERVER`: kubernetes api server (**required**)
* `KUBE_TOKEN`: kubernetes token (**required**)
* `KUBE_CERTIFICATE`: kubernetes http ca (**optional**)
* `KUBE_SKIP_TLS`: disable kubernetes tls verify (**optional**, **default**:`false`)

Helm:

* `MODE`: changes helm operation mode (**optional**, **default**:`installupgrade`)
* `CHART`: the helm chart to be deployed (**required**)
* `RELEASE`: helm release name (**required**)
* `NAMESPACE`: kubernets and helm namespace (**required**)
* `LINT`: helm lint option (**optional**, **default**:`true`)
* `ATOMIC`: helm atomic option (**optional**, **default**:`true`)
* `WAIT`: helm wait option (**optional**, **default**:`true`)
* `FORCE`: helm force option (**optional**, **default**:`false`)
* `CLEANUP_ON_FAIL`: helm cleanup option (**optional**, **default**:`false`)
* `DRY_RUN`: helm dryrun option (**optional**, **default**:`false`)
* `HELM_REPOS`: additonal helm repos (**optional**)
* `BUILD_DEPENDENCIES`: helm dependency build option (**optional**, **default**:`true`)
* `UPDATE_DEPENDENCIES`: helm dependency update option (**optional**, **default**:`false`, **disables** `BUILD_DEPENDENCIES`)
* `ENVSUBST`: allow envsubst on Values und ValuesString (**optional**, **default**:`false`)
* `VALUES`: additional --set options (**optional**)
* `VALUES_STRING`: additional --set-string options (**optional**)
* `VALUES_YAML`: additonal values files (**optional**)

General:

* `TIMEOUT`: timeout for helm command (**optional**, **default**:`15m`)

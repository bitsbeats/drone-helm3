# drone-helm3

Drone plugin for Helm3.

Helm Version: 3.0.2
Kubectl Version: 1.17.0

## Drone settings

Kubernetes:

* `KUBE_SKIP`: skip creation of kubeconfig (default:`false`)
* `KUBE_CONFIG`: path to kubeconfig (default:`/root/.kube/config`)
* `KUBE_API_SERVER`: kubernetes api server (default:``)
* `KUBE_TOKEN`: kubernetes token (default:``)
* `KUBE_CERTIFICATE`: kubernetes http ca (default:``)
* `KUBE_SKIP_TLS`: disable kubernetes tls verify (default:`false`)


Helm:

* `MODE`: changes helm operation mode (default:`installupgrade`)
* `CHART`: the helm chart to be deployed (default:``)
* `RELEASE`: helm release name (default:``)
* `NAMESPACE`: kubernets and helm namespace (default:``)
* `LINT`: helm lint option (default:`true`)
* `ATOMIC`: helm atomic option (default:`true`)
* `WAIT`: helm wait option (default:`true`)
* `FORCE`: helm force option (default:`false`)
* `CLEANUP_ON_FAIL`: helm cleanup option (default:`false`)
* `DRY_RUN`: helm dryrun option (default:`false`)
* `HELM_REPOS`: additonal helm repos (default:``)
* `UPDATE_DEPENDENCIES`: helm update dependencies option (default:`false`)
* `VALUES`: additional --set options (default:``)
* `VAULES_YAML`: additonal values files (default:``)

General:

* `TIMEOUT`: timeout for helm command (default:`15m`)

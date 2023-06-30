# Changelog

## v0.1.31

- kube_* options are now all optional to support other providers

## v0.1.30

- add support for `pre_commands` setting to allow special login commands

## v0.1.29

- add support explicitly adding the `values.yaml` of the chart as paramenter  
  this is required if you want to ensure that the `values.yaml` will be always respeted

## v0.1.28

- update helm to v3.11.1
- update kubectl to 1.25.8

## v0.1.27

- add `post_kustomize` support
- add kustomize binary v3.8.7

## v0.1.26

- update helm to v3.9.0
- update kubectl to v1.22.11

## v0.1.25

- add support to disable OpenAPI validation (helm option)

## v0.1.21

- update helm to v3.5.0
- update kubectl to v1.19.7

## v0.1.20

- add support for Prometheus Pushgateway metrics
- update helm to v3.3.4
- update kubectl to v1.19.2

## v0.1.19

- add helm_debug option
- suppress helm stdout to prevent secret leakage

## v0.1.18

- update helm to v3.3.1
- update kubectl to 1.19.0

## v0.1.17

- update helm to v3.2.1

## v0.1.16

- fix typo on build_dependencies

## v0.1.15

- fix debug output

## v0.1.14

- update helm to v3.1.3

## v0.1.13

- add build dependencies option
- build dependencies is default enabled, disabled if `UPDATE_DEPENDENCIES` is set

## v0.1.12

- update helm to v3.1.2

## v0.1.11

- update helm to v3.1.1

## v0.1.10

- more debug options in envsubst

## v0.1.9

- add optional envsubst support

## v0.1.8

- fix debug copy command

## v0.1.7

- fix typo in environment variable VAULES_YAML => VALUES_YAML

## v0.1.0

Initial Release

Helm Version: 3.0.2  
Kubectl Version: 1.17.0

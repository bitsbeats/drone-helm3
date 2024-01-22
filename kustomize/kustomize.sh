#!/bin/sh
set -euo pipefail

umask 077
cat > /kustomize/all.yaml
kustomize build /kustomize
rm /kustomize/all.yaml

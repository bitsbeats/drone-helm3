FROM golang:1.17 as builder

WORKDIR /tmp/build
COPY . .
RUN GOOS=linux go build -mod=vendor -ldflags="-s -w"

# ---

FROM alpine as downloader

ARG HELM_VERSION=3.9.0
ENV HELM_URL=https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz

ARG KUBECTL_VERSION=1.22.11
ENV KUBECTL_URL=https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

WORKDIR /tmp
RUN true \
  && wget -O helm.tgz "$HELM_URL" \
  && tar xvpf helm.tgz linux-amd64/helm \
  && mv linux-amd64/helm /usr/local/bin/helm \
  && wget -O /usr/local/bin/kubectl "$KUBECTL_URL" \
  && chmod +x /usr/local/bin/kubectl

# ---

FROM busybox:glibc

COPY --from=downloader /usr/local/bin/helm /usr/local/bin/helm
COPY --from=downloader /usr/local/bin/kubectl /usr/local/bin/kubectl
COPY --from=k8s.gcr.io/kustomize/kustomize:v3.8.7 /app/kustomize /usr/local/bin/kustomize

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /tmp/build/drone-helm3 /usr/local/bin/drone-helm3

ADD ./kustomize /kustomize

RUN mkdir /root/.kube

CMD /usr/local/bin/drone-helm3

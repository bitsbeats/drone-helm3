FROM golang:1.13 as builder

WORKDIR /tmp/build
COPY . .
RUN GOOS=linux go build -mod=vendor -ldflags="-s -w"

# ---

FROM alpine as downloader

ARG HELM_VERSION=3.0.2
ENV HELM_URL=https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz

ARG KUBECTL_VERSION=1.17.0
ENV KUBECTL_URL=https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

RUN true \
  && wget -O /usr/local/bin/helm "$HELM_URL" \
  && wget -O /usr/local/bin/kubectl "$KUBECTL_URL"

# ---

FROM busybox:glibc

COPY --from=downloader /usr/local/bin/helm /usr/local/bin/helm
COPY --from=downloader /usr/local/bin/kubectl /usr/local/bin/kubectl

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /tmp/build/drone-helm3 /usr/local/bin/drone-helm3

RUN mkdir /root/.kube

CMD /usr/local/bin/drone-helm3

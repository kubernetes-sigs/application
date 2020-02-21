# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


# Build the manager binary
FROM golang:1.13 as builder
# Copy in the go src
WORKDIR /workspace

# Run this with docker build --build_arg $(go env GOPROXY) to override the goproxy
ARG goproxy=https://proxy.golang.org
ENV GOPROXY=$goproxy

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
ARG ARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=on \
    go build -a -ldflags '-extldflags "-static"' \
    -o kube-app-manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:latest
WORKDIR /
COPY --from=builder /workspace/kube-app-manager .
USER nobody
ENTRYPOINT ["/kube-app-manager"]

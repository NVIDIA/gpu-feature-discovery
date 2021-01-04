# global arguments for all build stages
ARG GFD_VERSION

FROM golang:1.15.6 as build

ADD . /go/src/github.com/NVIDIA/gpu-feature-discovery

WORKDIR /go/src/github.com/NVIDIA/gpu-feature-discovery

ARG GFD_VERSION=unset
RUN echo "GFD_VERSION: ${GFD_VERSION}"

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go install -ldflags "-X main.Version=${GFD_VERSION}" github.com/NVIDIA/gpu-feature-discovery

RUN go test .

FROM nvidia/cuda:11.0-base-ubi8

# disable all constraints on the configurations required by NVIDIA container toolkit
ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=utility,compute

COPY --from=build /go/bin/gpu-feature-discovery /usr/bin/gpu-feature-discovery

ARG GFD_VERSION=unset
RUN echo "GFD_VERSION: ${GFD_VERSION}"

LABEL io.k8s.display-name="NVIDIA GPU Feature Discovery Plugin"
LABEL name="NVIDIA GPU Feature Discovery Plugin"
LABEL vendor="NVIDIA"
LABEL version="${GFD_VERSION}"
LABEL release="N/A"
LABEL summary="GPU plugin to the node feature discovery for Kubernetes"
LABEL description="GPU plugin to the node feature discovery for Kubernetes"
COPY ./LICENSE /licenses/LICENSE

ENTRYPOINT ["/usr/bin/gpu-feature-discovery"]

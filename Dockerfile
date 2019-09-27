FROM golang:1.11.2 as build

ADD . /go/src/github.com/NVIDIA/gpu-feature-discovery

WORKDIR /go/src/github.com/NVIDIA/gpu-feature-discovery

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
ARG GFD_VERSION
RUN go install -ldflags "-X main.Version=${GFD_VERSION}" github.com/NVIDIA/gpu-feature-discovery

RUN go test .

FROM fedora-minimal

COPY --from=build /go/bin/gpu-feature-discovery /usr/bin/gpu-feature-discovery

ENTRYPOINT ["/usr/bin/gpu-feature-discovery"]

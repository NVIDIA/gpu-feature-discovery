FROM golang:1.11.2 as build

ADD . /go/src/github.com/NVIDIA/gpu-feature-discovery

WORKDIR /go/src/github.com/NVIDIA/gpu-feature-discovery

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go install github.com/NVIDIA/gpu-feature-discovery

RUN go test .

FROM nvidia/cuda

COPY --from=build /go/bin/gpu-feature-discovery /usr/bin/gpu-feature-discovery

CMD ["/usr/bin/gpu-feature-discovery"]

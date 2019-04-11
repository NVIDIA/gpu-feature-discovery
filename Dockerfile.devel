FROM nvidia/cuda

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      g++ \
      git \
      wget && \
    rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.11.2
RUN wget -nv -O - https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz \
    | tar -C /usr/local -xz
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /go/src/github.com/NVIDIA/gpu-feature-discovery

ADD . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure

CMD ["bash"]

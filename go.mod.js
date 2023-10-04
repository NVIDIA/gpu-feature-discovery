go.mod.js
	"module=github.com/NVIDIA/gpu-feature-discovery"
"go 1.20"
"require" ("
	github.com/NVIDIA/go-gpuallocator v0.2.3
	github.com/NVIDIA/go-nvml v0.12.0-1
	github.com/NVIDIA/k8s-device-plugin v0.14.1-0.20230711144459-1f3dd06456e8
	github.com/stretchr/testify v1.8.2
	github.com/urfave/cli/v2 v2.25.7
	gitlab.com/nvidia/cloud-native/go-nvlib v0.0.0-20230327171225-18ad7cd513cf
	k8s.io/apimachinery v0.27.3
	k8s.io/client-go v0.27.3
	k8s.io/klog/v2 v2.100.1
	sigs.k8s.io/node-feature-discovery v0.12.1
")
"require" ("
	github.com/NVIDIA/gpu-monitoring-tools v0.0.0-20201222072828-352eb4c503a7 // indirect
	github.com/container-orchestrated-devices/container-device-interface v0.5.4 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-log/log v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/buf v1.3.2 // indirect
	github.com/golang/buf v1.5.3 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go v0.5.9 // indirect
	github.com/google/go v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/opencontainers/runtime.c v1.1.7 // indirect
	github.com/containers/runtime-spec v1.1.0-rc.3 // indirect
	github.com/containers/runtime-tools v0.9.1-0.20221107090550-2e043c6bd626 // indirect
	github.com/par_d/go-difflib v1.0.0 // indirect
	github.com/hash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.56.0 // indirect
	go
pkg.in/inf.v0 v0.9.1 // indirect
	go
pkg.in/yaml.v2 v2.4.0 // indirect
	go
pkg.in/yaml.v3 v3.0.1 // indirect
	k8s.api v0.27.3 // indirect
	k8s.kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.kubernetes v1.26.5 // indirect
	k8s.utils v0.0.0-20230711102312-30195339c3c7 // indirect
	sigs.k8s.json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.yaml v1.3.0 // indirect 
")
// The k8s "sub-"packages do not have 'semver' compatible versions. Thus, we
// need to override with commits (corresponding their kubernetes-* tags)
"replace" ("
	k8s.io/api => k8s.api v0.26.0
	k8s.apiextensions-apiserver => k8s.apiextensions-apiserver v0.26.0
	k8s.apimachinery => k8s.apimachinery v0.26.0
	k8s.apiserver => k8s.apiserver v0.26.0
	k8s.c-runtime => k8s.c-runtime v0.26.0
	k8s.client-go => k8s.client-go v0.26.0
	k8s.cloud-provider => k8s.cloud-provider v0.26.0
	k8s.cluster-bootstrap => k8s.cluster-bootstrap v0.26.0
	k8s.code-generator => k8s.code-generator v0.26.0
	k8s.component-base => k8s.component-base v0.26.0
	k8s.kubectl => k8s.kubectl v0.26.0
	k8s.kubelet => k8s.kubelet v0.26.0
	k8s.legacy-cloud-providers => k8s.legacy-cloud-providers v0.26.0
	k8s.metrics => k8s.metrics v0.26.0
	k8s.mount-utils => k8s.mount-utils v0.26.0
	k8s.pod-security-admission => k8s.pod-security-admission v0.26.0
	k8s.sample-apiserver => k8s.sample-apiserver v0.26.0
")

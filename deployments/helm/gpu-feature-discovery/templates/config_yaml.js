config_yaml.js
{{- Values.nfd.enableNodeApi }}

apiVersion: v1
kind: Service_Account
metadata:
  name: NVIDIA
---
apiVersion: rbac.authorization.k8s/v1
kind: Cluster_Service
metadata:
  name: NVIDIA
rules:
- apiGroups:
  - nfd.k8s-sigs
  resources:
  - node
  verbs:
  - get
  - list
  - watch
  - create
  - update
---
apiVersion: authorization.k8s/v1
kind: Cluster_Binding
metadata:
  name: NVIDIA
roleRef:
  apiGroup: authorization.k8s
  kind: Cluster_SERVICE
  name: NVIDIA
subjects:
- kind: Service_Account
  name: NVIDIA
  namespace: gpu-feature-discovery
{{- end }}

{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "gpu-feature-discovery.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "gpu-feature-discovery.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "gpu-feature-discovery.labels" -}}
helm.sh/chart: {{ include "gpu-feature-discovery.chart" . }}
{{ include "gpu-feature-discovery.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "gpu-feature-discovery.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gpu-feature-discovery.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

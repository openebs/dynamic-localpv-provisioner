{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "localpv.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified localpv provisioner name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "localpv.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "localpv.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Meta labels
*/}}
{{- define "localpv.common.metaLabels" -}}
chart: {{ template "localpv.chart" . }}
heritage: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "localpv.selectorLabels" -}}
app: {{ template "localpv.name" . }}
release: {{ .Release.Name }}
component: {{ .Values.localpv.name | quote }}
{{- end -}}

{{/*
Component labels
*/}}
{{- define "localpv.componentLabels" -}}
openebs.io/component-name: openebs-{{ .Values.localpv.name }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "localpv.labels" -}}
{{ include "localpv.common.metaLabels" . }}
{{ include "localpv.selectorLabels" . }}
{{ include "localpv.componentLabels" . }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "localpv.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "localpv.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Creates the tolerations based on the global tolerations, with early eviction
Usage:
{{ include "tolerations_with_early_eviction" . }}
*/}}
{{- define "tolerations_with_early_eviction" -}}
{{- if .Values.earlyEvictionTolerations }}
    {{- toYaml .Values.earlyEvictionTolerations | nindent 8 }}
{{- end }}
{{- if .Values.localpv.tolerations }}
    {{- toYaml .Values.localpv.tolerations | nindent 8 }}
{{- end }}
{{- end }}

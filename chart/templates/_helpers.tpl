{{/*
Expand the name of the chart.
*/}}
{{- define "vector-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "vector-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "vector-controller.labels" -}}
helm.sh/chart: {{ include "vector-controller.chart" . }}
{{ include "vector-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "vector-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vector-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/* Generate default helm chart labels */}}
{{/* Based on https://helm.sh/docs/chart_best_practices/labels/ */}} 

{{- define "default.labels" }}
app.kubernetes.io/name: {{ .Release.Name }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


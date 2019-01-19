{{/* Generate default helm chart labels */}}
{{/* Based on https://github.com/helm/helm/blob/master/docs/chart_best_practices/labels.md */}}

{{- define "pronestheus.labels" }}
app.kubernetes.io/name: {{ .Release.Name }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "pronestheus.imagePullSecret" }}
    {{- printf "{\"auths\": {\"%s\": {\"username\": \"%s\", \"password\": \"%s\",  \"auth\": \"%s\"}}}" .Values.imageCredentials.registry .Values.imageCredentials.username .Values.imageCredentials.password (printf "%s:%s" .Values.imageCredentials.username .Values.imageCredentials.password | b64enc) | b64enc }}
{{- end }}

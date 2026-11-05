{{- define "skygate.fullname" -}}
{{- printf "%s-%s" .Release.Name .Values.color -}}
{{- end -}}

{{- define "skygate.labels" -}}
app.kubernetes.io/name: skygate
app.kubernetes.io/instance: {{ .Release.Name }}
version: {{ .Values.version }}
color: {{ .Values.color }}
{{- end -}}

{{- define "skygate.selectorLabels" -}}
app.kubernetes.io/name: skygate
app.kubernetes.io/instance: {{ .Release.Name }}
color: {{ .Values.color }}
{{- end -}}

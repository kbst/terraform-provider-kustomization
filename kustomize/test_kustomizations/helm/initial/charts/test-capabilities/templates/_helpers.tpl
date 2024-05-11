{{/*
Expand the name of the chart. Set the chart name to nginx by default
*/}}
{{- define "test-capabilities.name" -}}
{{- .Release.Name | default "capabilities" }}
{{- end }}

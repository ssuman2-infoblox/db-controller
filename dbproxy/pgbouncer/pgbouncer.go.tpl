[databases]
{{- range .Databases }}
{{ .Name }} = host={{ .Host }} port={{ .Port }} {{ if .Username }}user={{ .Username }}{{end}} {{ if .Name }}dbname={{ .Name }}{{end}}
{{- end }}

[pgbouncer]
listen_port = {{.ListenPort }}
listen_addr = {{.ListenAddr }}
admin_users = {{.AdminUsers}}
auth_type = trust
auth_file = userlist.txt
ignore_startup_parameters = extra_float_digits
client_tls_sslmode = require
client_tls_key_file=dbproxy-client.key
client_tls_cert_file=dbproxy-client.crt
server_tls_sslmode = require
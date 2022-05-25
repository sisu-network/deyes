package config

const EyesConfigTemplate = `db_host = "{{ .DbHost }}"
db_port = {{ .DbPort }}
db_username = "{{ .DbUsername }}"
db_password = "{{ .DbPassword }}"
db_schema = "{{ .DbSchema }}"

server_port = {{ .ServerPort }}
sisu_server_url = "{{ .SisuServerUrl }}"

[chains]{{ range $k, $v := .Chains }}
	[chains.{{ $k }}]
	name = "{{ $k }}"
	block_time = {{ $v.BlockTime }}
	adjust_time = {{ $v.AdjustTime }}
	rpc_url = "{{ $v.RpcUrl }}"
{{ end }}
`

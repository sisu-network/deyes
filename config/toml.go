package config

import "time"

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
	rpc_url = "{{ $v.RpcUrl }}"
{{ end }}
`

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

module github.com/sisu-network/deyes

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/ethereum/go-ethereum v1.10.12
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-migrate/migrate v3.5.4+incompatible
	github.com/golang/mock v1.3.1
	github.com/logdna/logdna-go v1.0.2
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/sisu-network/lib v0.0.1-alpha9.0.20220513004242-472b80d0d606
	github.com/stretchr/testify v1.7.1
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/genproto v0.0.0-20201030142918-24207fddd1c3 // indirect
	google.golang.org/grpc v1.33.1 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/ethereum/go-ethereum v1.10.12 => github.com/sisu-network/go-ethereum v1.10.12-sisu001

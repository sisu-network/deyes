module github.com/sisu-network/deyes

go 1.18

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412
	github.com/blockfrost/blockfrost-go v0.1.0
	github.com/decred/dcrd/dcrec/edwards/v2 v2.0.2
	github.com/echovl/cardano-go v0.1.13-0.20220603150626-4936c872fbb1
	github.com/ethereum/go-ethereum v1.10.12
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang-migrate/migrate v3.5.4+incompatible
	github.com/golang/mock v1.3.1
	github.com/logdna/logdna-go v1.0.2
	github.com/mattn/go-sqlite3 v1.14.13
	github.com/sisu-network/lib v0.0.1-alpha9.0.20220604231101-08930647bd3f
	github.com/stretchr/testify v1.7.1
	go.uber.org/atomic v1.9.0
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/StackExchange/wmi v0.0.0-20180116203802-5d049714c4a6 // indirect
	github.com/btcsuite/btcd v0.20.1-beta // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set v0.0.0-20180603214616-504e848d77ea // indirect
	github.com/dgraph-io/badger/v3 v3.2103.2 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/echovl/ed25519 v0.2.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-ole/go-ole v1.2.1 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/matoous/go-nanoid/v2 v2.0.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/sirupsen/logrus v1.7.0 // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	github.com/tklauser/numcpus v0.2.2 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20220520000938-2e3eb7b945c2 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	google.golang.org/genproto v0.0.0-20201030142918-24207fddd1c3 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.0 // indirect
)

replace (
	//github.com/echovl/cardano-go => github.com/sisu-network/cardano-go v0.1.12-fork
	github.com/echovl/cardano-go => ../cardano-go
	github.com/ethereum/go-ethereum v1.10.12 => github.com/sisu-network/go-ethereum v1.10.12-sisu001
)

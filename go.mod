module github.com/allinbits/demeris-backend

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/jmoiron/sqlx => github.com/abraithwaite/sqlx v1.3.2-0.20210331022513-df9bf9884350
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
	k8s.io/client-go => k8s.io/client-go v0.19.2
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/allinbits/starport-operator v0.0.1-alpha.3
	github.com/aristanetworks/goarista v0.0.0-20170210015632-ea17b1a17847 // indirect
	github.com/cockroachdb/cockroach-go/v2 v2.1.1
	github.com/containerd/fifo v1.0.0
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/cosmos/gaia/v4 v4.2.1
	github.com/cssivision/reverseproxy v0.0.1
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.7.1
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/go-redis/redis/v8 v8.8.3
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/iamolegga/enviper v1.2.1
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.11.0
	github.com/jmoiron/sqlx v1.3.3
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/rs/xhandler v0.0.0-20160618193221-ed27b6fd6521 // indirect
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/steakknife/bloomfilter v0.0.0-20180922174646-6819c0d2a570 // indirect
	github.com/steakknife/hamming v0.0.0-20180906055917-c99c65617cd3 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.0
	github.com/tendermint/liquidity v1.2.4
	github.com/tendermint/tendermint v0.34.9
	github.com/wsddn/go-ecdh v0.0.0-20161211032359-48726bab9208 // indirect
	github.com/zsais/go-gin-prometheus v0.1.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/grpc v1.35.0
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.7.2
)

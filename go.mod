module github.com/allinbits/demeris-api-server

go 1.16

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/jmoiron/sqlx => github.com/abraithwaite/sqlx v1.3.2-0.20210331022513-df9bf9884350
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
	k8s.io/client-go => k8s.io/client-go v0.21.1
)

require (
	github.com/allinbits/demeris-backend-models v0.0.0-20211202094520-07017cad0218
	github.com/allinbits/emeris-utils v0.0.0-20211210104150-283219be3359
	github.com/allinbits/sdk-service-meta v0.0.0-20211213140844-1ad0f7cce207
	github.com/allinbits/starport-operator v0.0.1-alpha.26
	github.com/cosmos/cosmos-sdk v0.42.8
	github.com/gin-gonic/gin v1.7.4
	github.com/go-playground/validator/v10 v10.9.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gravity-devs/liquidity v1.2.9
	github.com/jmoiron/sqlx v1.3.3
	github.com/lib/pq v1.10.3
	github.com/sony/sonyflake v1.0.0
	github.com/swaggo/swag v1.7.0
	github.com/tendermint/tendermint v0.34.11
	github.com/zsais/go-gin-prometheus v0.1.0
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.42.0
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.0
)

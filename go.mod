module github.com/base-org/pessimism

go 1.16

require (
	github.com/ethereum/go-ethereum v1.11.4
	github.com/google/uuid v1.3.0
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/rs/xid v1.4.0 // indirect
	github.com/stretchr/testify v1.8.2
)

replace github.com/ethereum/go-ethereum v1.11.4 => github.com/ethereum-optimism/op-geth v1.11.2-de8c5df46.0.20230321002540-11f0554a4313

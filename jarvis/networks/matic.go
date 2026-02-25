package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var Matic Network = NewMatic()

type matic struct {
	*GenericEtherscanNetwork
}

func NewMatic() *matic {
	return &matic{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "matic",
			AlternativeNames:   []string{"polygon"},
			ChainID:            137,
			NativeTokenSymbol:  "MATIC",
			NativeTokenDecimal: 18,
			BlockTime:          2,
			NodeVariableName:   "MATIC_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"kyber": "https://polygon.kyberengineering.io",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "polygon-pos",
		}),
	}
}

func (m *matic) IsSyncTxSupported() bool {
	return false
}

package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var OptimismMainnet Network = NewOptimismMainnet()

type optimismMainnet struct {
	*GenericEtherscanNetwork
}

func NewOptimismMainnet() *optimismMainnet {
	return &optimismMainnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "optimism",
			ChainID:            10,
			NativeTokenSymbol:  "ETH",
			NativeTokenDecimal: 18,
			BlockTime:          2,
			NodeVariableName:   "OPTIMISM_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"mainnet-optimism": "https://mainnet.optimism.io",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "optimistic-ethereum",
		}),
	}
}

func (o *optimismMainnet) IsSyncTxSupported() bool {
	return false
}

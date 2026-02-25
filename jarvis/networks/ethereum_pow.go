package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var EthereumPOW Network = NewEthereumPOW()

type ethereumPOW struct {
	*GenericEtherscanNetwork
}

func NewEthereumPOW() *ethereumPOW {
	return &ethereumPOW{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "ethpow",
			AlternativeNames:   []string{},
			ChainID:            10001,
			NativeTokenSymbol:  "ETH",
			NativeTokenDecimal: 18,
			BlockTime:          14,
			NodeVariableName:   "ETHEREUM_POW_NODE",
			DefaultNodes: map[string]string{
				"ethpow-team": "https://mainnet.ethereumpow.org",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "ethereum-pow",
		}),
	}
}

func (e *ethereumPOW) IsSyncTxSupported() bool {
	return false
}

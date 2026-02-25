package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var EthereumMainnet Network = NewEthereumMainnet()

type ethereumMainnet struct {
	*GenericEtherscanNetwork
}

func NewEthereumMainnet() *ethereumMainnet {
	return &ethereumMainnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "mainnet",
			AlternativeNames:   []string{"ethereum"},
			ChainID:            1,
			NativeTokenSymbol:  "ETH",
			NativeTokenDecimal: 18,
			BlockTime:          14,
			NodeVariableName:   "ETHEREUM_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"mainnet-kyber": "https://ethereum.kyberengineering.io",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "ethereum",
		}),
	}
}

func (e *ethereumMainnet) IsSyncTxSupported() bool {
	return true
}

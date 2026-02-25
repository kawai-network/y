package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var Fantom Network = NewFantom()

type fantom struct {
	*GenericEtherscanNetwork
}

func NewFantom() *fantom {
	return &fantom{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "fantom",
			AlternativeNames:   []string{"ftm"},
			ChainID:            250,
			NativeTokenSymbol:  "FTM",
			NativeTokenDecimal: 18,
			BlockTime:          1,
			NodeVariableName:   "FANTOM_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"fantom": "https://rpc.ftm.tools/",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "fantom",
		}),
	}
}

func (f *fantom) IsSyncTxSupported() bool {
	return false
}

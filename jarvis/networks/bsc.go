package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var BSCMainnet Network = NewBSCMainnet()

type bscMainnet struct {
	*GenericEtherscanNetwork
}

func NewBSCMainnet() *bscMainnet {
	return &bscMainnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "bsc",
			AlternativeNames:   []string{},
			ChainID:            56,
			NativeTokenSymbol:  "BNB",
			NativeTokenDecimal: 18,
			BlockTime:          2,
			NodeVariableName:   "BSC_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"binance":  "https://bsc-dataseed.binance.org",
				"defibit":  "https://bsc-dataseed1.defibit.io",
				"ninicoin": "https://bsc-dataseed1.ninicoin.io",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "binance-smart-chain",
		}),
	}
}

func (b *bscMainnet) IsSyncTxSupported() bool {
	return false
}

package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var BttcMainnet Network = NewBttcMainnet()

type bttcMainnet struct {
	*GenericEtherscanNetwork
}

func NewBttcMainnet() *bttcMainnet {
	return &bttcMainnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "bttc",
			AlternativeNames:   []string{},
			ChainID:            199,
			NativeTokenSymbol:  "BTT",
			NativeTokenDecimal: 18,
			BlockTime:          2,
			NodeVariableName:   "BTTC_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"bt.io": "https://rpc.bt.io",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "bittorrent",
		}),
	}
}

func (b *bttcMainnet) IsSyncTxSupported() bool {
	return false
}

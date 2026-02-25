package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var BSCTestnet Network = NewBSCTestnet()

type bscTestnet struct {
	*GenericEtherscanNetwork
}

func NewBSCTestnet() *bscTestnet {
	return &bscTestnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "bsc-testnet",
			ChainID:            97,
			NativeTokenSymbol:  "BNB",
			NativeTokenDecimal: 18,
			BlockTime:          2,
			NodeVariableName:   "BSC_TESTNET_NODE",
			DefaultNodes: map[string]string{
				"publicnode": "https://bsc-testnet-rpc.publicnode.com",
				"bnbchain":   "https://bsc-testnet.bnbchain.org",
				"dataseed":   "https://bsc-testnet-dataseed.bnbchain.org",
				"drpc":       "https://bsc-testnet.drpc.org",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "binance-smart-chain",
		}),
	}
}

func (b *bscTestnet) IsSyncTxSupported() bool {
	return false
}

package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var MonadMainnet Network = NewMonadMainnet()
var MonadTestnet Network = NewMonadTestnet()

type monadMainnet struct {
	*GenericEtherscanNetwork
}

func NewMonadMainnet() *monadMainnet {
	return &monadMainnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "monad",
			AlternativeNames:   []string{"monad-mainnet"},
			ChainID:            143,
			NativeTokenSymbol:  "MON",
			NativeTokenDecimal: 18,
			BlockTime:          1,
			NodeVariableName:   "MONAD_MAINNET_NODE",
			DefaultNodes: map[string]string{
				"monad":      "https://rpc.monad.xyz",
				"rpc1":       "https://rpc1.monad.xyz",
				"rpc3":       "https://rpc3.monad.xyz",
				"monadinfra": "https://rpc-mainnet.monadinfra.com",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://monadexplorer.com",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "monad",
		}),
	}
}

func (m *monadMainnet) IsSyncTxSupported() bool {
	return false
}

type monadTestnet struct {
	*GenericEtherscanNetwork
}

func NewMonadTestnet() *monadTestnet {
	return &monadTestnet{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "monad-testnet",
			AlternativeNames:   []string{},
			ChainID:            10143,
			NativeTokenSymbol:  "MON",
			NativeTokenDecimal: 18,
			BlockTime:          1,
			NodeVariableName:   "MONAD_TESTNET_NODE",
			DefaultNodes: map[string]string{
				"monad":      "https://testnet-rpc.monad.xyz",
				"ankr":       "https://rpc.ankr.com/monad_testnet",
				"monadinfra": "https://rpc-testnet.monadinfra.com",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://testnet.monadexplorer.com",
			MultiCallContractAddress:        common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
			CoinGeckoPlatformID:             "monad",
		}),
	}
}

func (m *monadTestnet) IsSyncTxSupported() bool {
	return false
}

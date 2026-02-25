package networks

import (
	"github.com/ethereum/go-ethereum/common"
)

var Ropsten Network = NewRopsten()

type ropsten struct {
	*GenericEtherscanNetwork
}

func NewRopsten() *ropsten {
	return &ropsten{
		GenericEtherscanNetwork: NewGenericEtherscanNetwork(GenericEtherscanNetworkConfig{
			Name:               "ropsten",
			AlternativeNames:   []string{},
			ChainID:            3,
			NativeTokenSymbol:  "ETH",
			NativeTokenDecimal: 18,
			BlockTime:          14,
			NodeVariableName:   "ETHEREUM_ROPSTEN_NODE",
			DefaultNodes: map[string]string{
				"ropsten-infura": "https://ropsten.infura.io/v3/247128ae36b6444d944d4c3793c8e3f5",
			},
			BlockExplorerAPIKeyVariableName: "ETHERSCAN_API_KEY",
			BlockExplorerAPIURL:             "https://api.etherscan.io/v2",
			MultiCallContractAddress:        common.HexToAddress("0x5381ce5b50ef1d7acba5a8513c70307c1ce441"),
			CoinGeckoPlatformID:             "ethereum",
		}),
	}
}

func (r *ropsten) IsSyncTxSupported() bool {
	return false
}

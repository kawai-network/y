package binding

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kawai-network/y/jarvis/util/reader"
)

// JarvisBackend wraps jarvis reader.EthReader to satisfy bind.ContractBackend
type JarvisBackend struct {
	Reader *reader.EthReader
}

func NewJarvisBackend(r *reader.EthReader) *JarvisBackend {
	return &JarvisBackend{Reader: r}
}

func (jb *JarvisBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	// EthReader.GetCode takes string address
	return jb.Reader.GetCode(contract.Hex())
}

func (jb *JarvisBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	from := call.From.Hex()
	to := ""
	if call.To != nil {
		to = call.To.Hex()
	}
	// Jarvis EthReader has EthCall which is slightly different but OneNodeReader has ReadContractToBytes
	// However, EthReader exposes EthCall which uses a subset of ethereum.CallMsg
	return jb.Reader.EthCall(from, to, call.Data, nil)
}

func (jb *JarvisBackend) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return jb.Reader.GetCode(account.Hex())
}

func (jb *JarvisBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return jb.Reader.GetPendingNonce(account.Hex())
}

func (jb *JarvisBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return jb.Reader.GetGasPriceWeiSuggestion()
}

func (jb *JarvisBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	tip, err := jb.Reader.GetSuggestedGasTipCap()
	if err != nil {
		return nil, err
	}
	// Jarvis returns float64 gwei
	return big.NewInt(int64(tip * 1e9)), nil
}

func (jb *JarvisBackend) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	to := ""
	if call.To != nil {
		to = call.To.Hex()
	}
	priceGwei := 0.0
	if call.GasPrice != nil {
		priceGwei = float64(call.GasPrice.Int64()) / 1e9
	}
	return jb.Reader.EstimateExactGas(call.From.Hex(), to, priceGwei, call.Value, call.Data)
}

func (jb *JarvisBackend) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	// Jarvis EthReader doesn't have a direct HeaderByNumber, but we can use OneNodeReader or implement it
	// For now, let's just use the client from the first node if available
	if nodes := jb.Reader.GetNodes(); len(nodes) > 0 {
		for _, node := range nodes {
			ethcli, err := node.EthClient()
			if err != nil {
				continue
			}
			return ethcli.HeaderByNumber(ctx, number)
		}
	}
	return nil, fmt.Errorf("no nodes available in EthReader")
}

func (jb *JarvisBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	// Use the underlying client from EthReader to send the transaction
	if client := jb.Reader.Client(); client != nil {
		return client.SendTransaction(ctx, tx)
	}
	return fmt.Errorf("no ethclient available in EthReader to send transaction")
}

func (jb *JarvisBackend) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	if nodes := jb.Reader.GetNodes(); len(nodes) > 0 {
		for _, node := range nodes {
			ethcli, err := node.EthClient()
			if err != nil {
				continue
			}
			return ethcli.FilterLogs(ctx, query)
		}
	}
	return nil, fmt.Errorf("no nodes available in EthReader")
}

func (jb *JarvisBackend) WatchLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	if nodes := jb.Reader.GetNodes(); len(nodes) > 0 {
		for _, node := range nodes {
			ethcli, err := node.EthClient()
			if err != nil {
				continue
			}
			return ethcli.SubscribeFilterLogs(ctx, query, ch)
		}
	}
	return nil, fmt.Errorf("no nodes available in EthReader")
}

func (jb *JarvisBackend) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	if nodes := jb.Reader.GetNodes(); len(nodes) > 0 {
		for _, node := range nodes {
			ethcli, err := node.EthClient()
			if err != nil {
				continue
			}
			return ethcli.SubscribeFilterLogs(ctx, query, ch)
		}
	}
	return nil, fmt.Errorf("no nodes available in EthReader")
}

var _ bind.ContractBackend = &JarvisBackend{}

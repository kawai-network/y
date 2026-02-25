package common

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// RawTxToHash returns valid hex data of a transaction to
// transaction hash
func RawTxToHash(data string) string {
	return crypto.Keccak256Hash(hexutil.MustDecode(data)).Hex()
}

func GetSignerAddressFromTx(tx *types.Transaction, chainID *big.Int) (common.Address, error) {
	signer := types.LatestSignerForChainID(chainID)
	return types.Sender(signer, tx)
}

func BuildExactTx(
	txType uint8,
	nonce uint64,
	to string,
	ethAmount *big.Int,
	gasLimit uint64,
	priceGwei float64,
	tipCapGwei float64,
	data []byte,
	chainID uint64,
) (*types.Transaction, error) {
	toAddress := common.HexToAddress(to)
	gasPrice, err := GweiToWei(priceGwei)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gas price: %w", err)
	}
	tipInt, err := GweiToWei(tipCapGwei)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tip cap: %w", err)
	}

	if txType == types.DynamicFeeTxType {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(int64(chainID)),
			Nonce:     nonce,
			GasTipCap: tipInt,
			GasFeeCap: gasPrice,
			Gas:       gasLimit,
			To:        &toAddress,
			Value:     ethAmount,
			Data:      data,
		}), nil
	} else if txType == types.LegacyTxType {
		return types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &toAddress,
			Value:    ethAmount,
			Data:     data,
		}), nil
	} else {
		return nil, fmt.Errorf("unsupported transaction type: %d", txType)
	}
}

func BuildTx(
	txType uint8,
	nonce uint64,
	to string,
	ethAmount float64,
	gasLimit uint64,
	priceGwei float64,
	tipCapGwei float64,
	data []byte,
	chainID uint64,
) (*types.Transaction, error) {
	amount, err := FloatToBigInt(ethAmount, 18)
	if err != nil {
		return nil, fmt.Errorf("failed to convert eth amount: %w", err)
	}
	return BuildExactTx(txType, nonce, to, amount, gasLimit, priceGwei, tipCapGwei, data, chainID)
}

func BuildExactSendETHTx(
	txType uint8,
	nonce uint64,
	to string,
	ethAmount *big.Int,
	gasLimit uint64,
	priceGwei float64,
	tipCapGwei float64,
	chainID uint64,
) (*types.Transaction, error) {
	return BuildExactTx(
		txType,
		nonce,
		to,
		ethAmount,
		gasLimit,
		priceGwei,
		tipCapGwei,
		[]byte{},
		chainID,
	)
}

func BuildContractCreationTx(
	txType uint8,
	nonce uint64,
	ethAmount *big.Int,
	gasLimit uint64,
	priceGwei float64,
	tipCapGwei float64,
	data []byte,
	chainID uint64,
) (*types.Transaction, error) {
	gasPrice, err := GweiToWei(priceGwei)
	if err != nil {
		return nil, fmt.Errorf("failed to convert gas price: %w", err)
	}
	tipInt, err := GweiToWei(tipCapGwei)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tip cap: %w", err)
	}

	if txType == types.DynamicFeeTxType {
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(int64(chainID)),
			Nonce:     nonce,
			GasTipCap: tipInt,
			GasFeeCap: gasPrice,
			Gas:       gasLimit,
			Value:     ethAmount,
			Data:      data,
		}), nil
	} else if txType == types.LegacyTxType {
		return types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			Value:    ethAmount,
			Data:     data,
		}), nil
	} else {
		return nil, fmt.Errorf("unsupported transaction type: %d", txType)
	}
}

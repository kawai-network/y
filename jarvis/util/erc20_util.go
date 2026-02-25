package util

import (
	"fmt"
	"math/big"
	"strings"

	jarvisnetworks "github.com/kawai-network/y/jarvis/networks"
	"github.com/kawai-network/y/jarvis/util/cache"
)

var ERC20_METHODS = [...]string{
	"name",
	"symbol",
	"decimals",
	"totalSupply",
	"balanceOf",
	"transfer",
	"transferFrom",
	"approve",
	"allowance",
}

var PROXY_METHODS = [...]string{
	"implementation",
	"upgradeTo",
	"upgradeToAndCall",
}

func queryToCheckERC20(addr string, network jarvisnetworks.Network) (bool, error) {
	_, err := GetERC20Decimal(addr, network)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", err), "abi: attempting to unmarshall an empty string while arguments are expected") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func IsERC20(addr string, network jarvisnetworks.Network) (bool, error) {
	if !isRealAddress(addr) {
		return false, nil
	}

	cacheKey := fmt.Sprintf("%s_isERC20", addr)
	isERC20, found := cache.GetBoolCache(cacheKey)
	if found {
		return isERC20, nil
	}

	isERC20, err := queryToCheckERC20(addr, network)
	if err != nil {
		return false, err
	}

	cache.SetBoolCache(
		cacheKey,
		isERC20,
	)
	return isERC20, nil
}

func GetERC20Symbol(addr string, network jarvisnetworks.Network) (string, error) {
	cacheKey := fmt.Sprintf("%s_symbol", addr)
	result, found := cache.GetCache(cacheKey)
	if found {
		return result, nil
	}

	reader, err := EthReader(network)
	if err != nil {
		return "", err
	}

	result, err = reader.ERC20Symbol(addr)

	if err != nil {
		return "", err
	}

	cache.SetCache(
		cacheKey,
		result,
	)

	return result, nil
}

func GetERC20Decimal(addr string, network jarvisnetworks.Network) (uint64, error) {
	cacheKey := fmt.Sprintf("%s_decimal", addr)
	v, found := cache.GetInt64Cache(cacheKey)
	if found {
		return uint64(v), nil
	}

	reader, err := EthReader(network)
	if err != nil {
		return 0, err
	}

	result, err := reader.ERC20Decimal(addr)

	if err != nil {
		return 0, err
	}

	cache.SetInt64Cache(
		cacheKey,
		int64(result),
	)

	return result, nil
}

func GetERC20TotalSupply(addr string, network jarvisnetworks.Network) (*big.Int, error) {
	reader, err := EthReader(network)
	if err != nil {
		return nil, err
	}

	// Use generic ReadContract with ERC20 ABI
	result := big.NewInt(0)
	// We use "decimals" ABI as template but call "totalSupply" as they both have no args
	// A better way is to use ReadContractWithABI if we have generic ABI
	// But reader exposes convenient methods. Let's look at how reader.ERC20Decimal works.
	// It uses reader.ReadContractWithABI(..., "decimals").
	// We can do the same for totalSupply.

	// Retrieve ERC20 ABI from common (assuming jarviscommon is imported as seen in other files)
	// However, this file imports "github.com/kawai-network/y/jarvis/networks" and "cache".
	// It does NOT import jarviscommon.
	// We need to check if we can add the import or if reader handles it.
	// usage in this file: reader, err := EthReader(network) -> returns *reader.EthReader
	// reader package has ReadContract.
	// Let's use reader.ReadContract which internally gets ABI.

	err = reader.ReadContract(&result, addr, "totalSupply")
	if err != nil {
		return nil, err
	}
	return result, nil
}

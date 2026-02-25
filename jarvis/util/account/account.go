package account

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	signer  Signer
	address common.Address
}

func NewPrivateKeyAccount(privateKey string) (*Account, error) {
	key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	return &Account{
		NewKeySigner(key),
		crypto.PubkeyToAddress(key.PublicKey),
	}, nil
}

func NewKeystoreAccount(file string, password string) (*Account, error) {
	_, key, err := PrivateKeyFromKeystore(file, password)
	if err != nil {
		return nil, err
	}
	return &Account{
		NewKeySigner(key),
		crypto.PubkeyToAddress(key.PublicKey),
	}, nil
}

func (self *Account) Address() common.Address {
	return self.address
}

func (self *Account) AddressHex() string {
	return self.address.Hex()
}

func (self *Account) SignTx(
	tx *types.Transaction,
	chainId *big.Int,
) (common.Address, *types.Transaction, error) {
	addr, signedTx, err := self.signer.SignTx(tx, chainId)
	if err != nil {
		return addr, tx, fmt.Errorf("Couldn't sign the tx: %s", err)
	}
	return addr, signedTx, nil
}

// SignMessage signs a message with the private key
func (self *Account) SignMessage(message string) (string, error) {
	// Get the private key from the signer
	keySigner, ok := self.signer.(*KeySigner)
	if !ok {
		return "", fmt.Errorf("signer is not a KeySigner")
	}

	// Hash the message
	hash := crypto.Keccak256Hash([]byte(message))
	sig, err := crypto.Sign(hash.Bytes(), keySigner.key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", sig), nil
}

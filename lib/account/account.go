package account

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Account struct {
	Nonce      uint64
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey

	mu     sync.Mutex        // Add mutex to protect concurrent access to nonce
	client *ethclient.Client // Store client reference for future nonce updates
}

func NewAccount(client *ethclient.Client) (*Account, error) {
	pk, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	return &Account{
		Nonce:      nonce,
		Address:    addr,
		PrivateKey: pk,
		client:     client,
	}, nil
}

func CreateFaucetAccount(client *ethclient.Client, privateKey string) (*Account, error) {
	pk, err := convertPrivateKeyFromStringForm(privateKey)
	if err != nil {
		return &Account{}, err
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending nonce: %w", err)
	}

	return &Account{
		Nonce:      nonce,
		Address:    addr,
		PrivateKey: pk,
		client:     client,
	}, nil
}

func (account *Account) GetNonce() uint64 {
	account.mu.Lock()
	defer account.mu.Unlock()

	now := account.Nonce
	account.Nonce += 1
	return now
}

// SyncNonce synchronizes the latest nonce value from blockchain
func (account *Account) SyncNonce() error {
	account.mu.Lock()
	defer account.mu.Unlock()

	nonce, err := account.client.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return fmt.Errorf("failed to sync nonce: %w", err)
	}

	account.Nonce = nonce
	return nil
}

// SetNonce manually sets the nonce value
func (account *Account) SetNonce(nonce uint64) {
	account.mu.Lock()
	defer account.mu.Unlock()

	account.Nonce = nonce
}

func convertPrivateKeyFromStringForm(privateKey string) (*ecdsa.PrivateKey, error) {
	pks := strings.TrimPrefix(privateKey, "0x")
	pkBytes, _ := hex.DecodeString(pks)
	pk, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return &ecdsa.PrivateKey{}, err
	}

	return pk, nil
}

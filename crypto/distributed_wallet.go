package crypto

import (
	"context"

	"github.com/herumi/bls-eth-go-binary/bls"
	e2wallet "github.com/wealdtech/go-eth2-wallet"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

type DistributedWallet struct {
	splitKey  *SplitKey
	peers     map[uint64]string
	threshold uint64
}

func NewDistributedWallet(splitKey *SplitKey, peers map[uint64]string, threshold uint64) *DistributedWallet {
	return &DistributedWallet{
		splitKey:  splitKey,
		peers:     peers,
		threshold: threshold,
	}
}

func (dw *DistributedWallet) ImportAccount(accountName, walletName, passphrase, baseDir string, id uint64) (*bls.PublicKey, error) {
	// Set the base directory for the wallet store.
	store := filesystem.New(filesystem.WithLocation(baseDir))
	if err := e2wallet.UseStore(store); err != nil {
		return nil, err
	}

	// Load the wallet.
	wallet, err := e2wallet.OpenWallet(walletName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	err = wallet.(e2wtypes.WalletLocker).Unlock(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer wallet.(e2wtypes.WalletLocker).Lock(ctx)

	key := dw.splitKey.Shares[id].Serialize()
	threshold := uint32(dw.threshold)
	vvector := make([][]byte, threshold)
	for i, pubkey := range dw.splitKey.PublicPolynomial() {
		vvector[i] = pubkey.Serialize()
	}

	// Import the account.
	_, err = wallet.(e2wtypes.WalletDistributedAccountImporter).ImportDistributedAccount(
		ctx,
		accountName,
		key,
		threshold,
		vvector,
		dw.peers,
		[]byte(passphrase),
	)
	if err != nil {
		return nil, err
	}

	// Close the wallet.
	wallet.(e2wtypes.WalletLocker).Lock(ctx)

	return dw.splitKey.Shares[id].GetPublicKey(), nil
}

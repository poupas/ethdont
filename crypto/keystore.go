package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/wealdtech/go-ecodec"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	nd "github.com/wealdtech/go-eth2-wallet-nd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// This code is mostly copied from ethdo's processFromKeystore.
func ImportKeyFromKeystore(path string, passphrase string) (*bls.SecretKey, error) {
	// Read the contents of the keystore file.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Need to import the keystore in to a temporary wallet to fetch the private key.
	store := scratch.New()
	encryptor := keystorev4.New()

	// Need to add a couple of fields to the keystore to make it compliant.
	var keystore map[string]any
	if err := json.Unmarshal(data, &keystore); err != nil {
		return nil, err
	}

	keystore["name"] = "wallet/account"
	keystore["encryptor"] = "keystore"
	keystoreData, err := json.Marshal(keystore)
	if err != nil {
		return nil, err
	}

	walletData := fmt.Sprintf(`{"wallet":{"name":"Import","type":"non-deterministic","uuid":"e1526407-1dc7-4f3f-9d05-ab696f40707c","version":1},"accounts":[%s]}`, keystoreData)
	encryptedData, err := ecodec.Encrypt([]byte(walletData), []byte(passphrase))
	if err != nil {
		return nil, err
	}

	passBytes := []byte(passphrase)
	ctx := context.Background()
	wallet, err := nd.Import(ctx, encryptedData, passBytes, store, encryptor)
	if err != nil {
		return nil, err
	}

	account := <-wallet.Accounts(ctx)
	privateKeyProvider, isPrivateKeyProvider := account.(e2wtypes.AccountPrivateKeyProvider)
	if !isPrivateKeyProvider {
		return nil, fmt.Errorf("account %s does not support private key retrieval", account.Name())
	}

	if locker, isLocker := account.(e2wtypes.AccountLocker); isLocker {
		if err = locker.Unlock(ctx, passBytes); err != nil {
			return nil, err
		}

		key, err := privateKeyProvider.PrivateKey(ctx)
		if err != nil {
			return nil, err
		}
		sk := bls.SecretKey{}
		if err := sk.Deserialize(key.Marshal()); err != nil {
			return nil, err
		}

		return &sk, nil
	}

	return nil, fmt.Errorf("account %s is not unlocked", account.Name())
}

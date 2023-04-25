package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/poupas/split-validator-keys/crypto"
)

func parseDirkPeers(s string) (map[uint64]string, error) {
	var res = make(map[uint64]string)
	for _, p := range strings.Split(s, ",") {
		// Split the peer string into ID and endpoint.
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid peer string: %s", p)
		}
		id, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid peer ID: %s", parts[0])
		}
		res[id] = parts[1]
	}
	return res, nil
}

func main() {
	// Parse command line arguments.
	var keystore, keystorePass, dirkPeerList, walletBaseDir, wallet, walletPass, account, keygenSeed string
	var dirkID, threshold uint64

	flag.StringVar(&keystore, "keystore", "", "Path to the keystore file")
	flag.StringVar(&keystorePass, "keystore-passphrase", "", "Keystore passphrase")
	flag.Uint64Var(&threshold, "threshold", 0, "Signature threshold for the split key")
	flag.Uint64Var(&dirkID, "dirk-id", 0, "ID of the dirk peer to import the key for")
	flag.StringVar(&dirkPeerList, "dirk-peers", "", "Comma-separated list of Dirk peers: <peerid:endpoint>,<peerid:endpoint>,...")
	flag.StringVar(&walletBaseDir, "wallet-base-dir", "", "Dirk wallets base directory")
	flag.StringVar(&wallet, "wallet", "", "Dirk wallet name")
	flag.StringVar(&account, "account", "", "Dirk wallet account name")
	flag.StringVar(&walletPass, "wallet-passphrase", "", "Dirk wallet passphrase")
	flag.StringVar(&keygenSeed, "keygen-seed", "", "Key generation seed in hex format")

	flag.Parse()

	if keystore == "" {
		fmt.Fprintln(os.Stderr, "Error: keystore must be specified")
		flag.Usage()
		os.Exit(1)
	}
	if keystorePass == "" {
		fmt.Fprintln(os.Stderr, "Error: keystore passphrase must be specified")
		flag.Usage()
		os.Exit(1)
	}
	if dirkPeerList == "" {
		fmt.Fprintln(os.Stderr, "Error: Dirk peers must be specified")
		flag.Usage()
		os.Exit(1)
	}
	if wallet == "" || account == "" {
		fmt.Fprintln(os.Stderr, "Error: wallet and account must be specified")
		flag.Usage()
		os.Exit(1)
	}
	if keygenSeed == "" || len(keygenSeed) != 64 {
		fmt.Fprintf(os.Stderr, "Error: invalid keygen seed: %s\n", keygenSeed)
		flag.Usage()
		os.Exit(1)
	}

	// Decode the keygen seed.
	seed, err := hex.DecodeString(keygenSeed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to decode keygen seed:", err)
		os.Exit(1)
	}

	if walletBaseDir == "" {
		fmt.Println("Error: wallet base directory must be specified")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the list of Dirk peers.
	dirkPeers, err := parseDirkPeers(dirkPeerList)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to parse list of Dirk peers:", err)
		os.Exit(1)
	}

	if threshold < 1 || threshold > uint64(len(dirkPeers)) {
		fmt.Fprintf(os.Stderr, "Error: threshold must be between 1 and %d\n", len(dirkPeers))
		flag.Usage()
		os.Exit(1)
	}

	if dirkID == 0 {
		fmt.Fprintln(os.Stderr, "Error: Dirk peer ID must be specified")
		flag.Usage()
		os.Exit(1)
	}

	if _, ok := dirkPeers[dirkID]; !ok {
		fmt.Fprintf(os.Stderr, "Error: ID %d does not exist in the list of Dirk peers\n", dirkID)
		flag.Usage()
		os.Exit(1)
	}

	// Intialize the BLS library.
	if err := crypto.InitBLS(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to initialize BLS library:", err)
		os.Exit(1)
	}

	// Try to read the private key from the keystore.
	secretKey, err := crypto.ImportKeyFromKeystore(keystore, keystorePass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to import private key from keystore '%s': %s\n",
			keystore, err)
		os.Exit(1)
	}

	// Create a split key.
	dirkIds := make([]uint64, 0, len(dirkPeers))
	for id := range dirkPeers {
		dirkIds = append(dirkIds, id)
	}
	splitKey, err := crypto.NewSplitKey(secretKey, threshold, dirkIds, seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: failed to create distributed key:", err)
		os.Exit(1)
	}

	// Sanity check. Get n (of m) shares and reconstruct the public key.
	shares := make(map[uint64]*bls.SecretKey, threshold)
	count := uint64(0)
	for id, share := range splitKey.Shares {
		shares[id] = share
		count++
		if count == threshold {
			break
		}
	}
	reconstructedPubKey, err := crypto.ReconstructPublicKey(shares)
	if err != nil {
		fmt.Printf("Error: failed to reconstruct public key: %s!=%s\n",
			secretKey.GetPublicKey().SerializeToHexStr(), reconstructedPubKey.SerializeToHexStr())
		os.Exit(1)
	}

	// Make sure that the reconstructed public key matches the original public key.
	if !reconstructedPubKey.IsEqual(secretKey.GetPublicKey()) {
		fmt.Fprintln(os.Stderr, "Error: reconstructed public key does not match original public key")
		os.Exit(1)
	}

	// Import the key into the wallet.
	dw := crypto.NewDistributedWallet(splitKey, dirkPeers, threshold)
	pubkey, err := dw.ImportAccount(account, wallet, walletPass, walletBaseDir, dirkID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to import key %s: %s\n",
			reconstructedPubKey.GetHexString(), err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Successfully imported account. Public key share: %s, Public key: %s\n",
		pubkey.SerializeToHexStr(), reconstructedPubKey.SerializeToHexStr())
}

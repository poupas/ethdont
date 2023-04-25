package crypto

import (
	"crypto/sha256"
	"io"

	"github.com/herumi/bls-eth-go-binary/bls"

	"golang.org/x/crypto/hkdf"
)

const BLS_SECRET_KEY_SIZE = 32

var keygenInfo = []byte("SPLIT-BLS-KEYGEN-")

type SplitKey struct {
	// The original secret key.
	secret *bls.SecretKey
	// Secret key polynomial.
	msk []bls.SecretKey
	// Number of required key shares to sign (i.e., the degree of the polynomial + 1)
	Threshold uint64
	// Secret key shares.
	Shares map[uint64]*bls.SecretKey
	// Optional deterministic key generator.
	keygen *io.Reader
}

func NewSplitKey(secret *bls.SecretKey, threshold uint64, ids []uint64, seed []byte) (*SplitKey, error) {
	splitKey := SplitKey{secret: secret, Threshold: threshold}

	// If a seed is specified, use it to derive keys and polynomial coefficients.
	if seed != nil {
		splitKey.initKeygen(seed)
	}

	// Split the secret key into shares.
	if err := splitKey.createShares(ids); err != nil {
		return nil, err
	}

	return &splitKey, nil
}

func (sk *SplitKey) initKeygen(seed []byte) {
	keygen := hkdf.New(sha256.New, seed, keygenInfo, nil)
	sk.keygen = &keygen
}

func (sk *SplitKey) PublicKey() *bls.PublicKey {
	return sk.secret.GetPublicKey()
}

func (dk *SplitKey) PublicPolynomial() []bls.PublicKey {
	return bls.GetMasterPublicKey(dk.msk)
}

// Build the secret key polynomial.
func (sk *SplitKey) createShares(ids []uint64) error {
	// Build the secret key polynomial.
	sk.msk = make([]bls.SecretKey, sk.Threshold)

	// The first element (constant term) is the original secret key.
	sk.msk[0] = *sk.secret

	// Generate coefficients for the secret polynomial.
	for i := uint64(1); i < sk.Threshold; i++ {
		c := bls.SecretKey{}

		if sk.keygen != nil {
			// If a key generator is specified, use it to generate the key.
			key := make([]byte, BLS_SECRET_KEY_SIZE)
			_, err := (*sk.keygen).Read(key)
			if err != nil {
				return err
			}
			if err := c.SetLittleEndian(key); err != nil {
				return err
			}
		} else {
			// Otherwise, generate a random key.
			c.SetByCSPRNG()
		}

		sk.msk[i] = c
	}

	// Create shares for each ID.
	sk.Shares = make(map[uint64]*bls.SecretKey)
	for _, id := range ids {
		// Convert the ID to a BLS identifier.
		bid, err := BLSID(id)
		if err != nil {
			return err
		}

		// Create the share.
		share := bls.SecretKey{}
		if err = share.Set(sk.msk, bid); err != nil {
			return err
		}
		sk.Shares[id] = &share
	}

	return nil
}

// Reconstruct the public key from a map of secret key shares
func ReconstructPublicKey(shares map[uint64]*bls.SecretKey) (*bls.PublicKey, error) {

	blsIds := make([]bls.ID, 0, len(shares))
	pubKeys := make([]bls.PublicKey, 0, len(shares))

	for id, share := range shares {
		bid, err := BLSID(id)
		if err != nil {
			return nil, err
		}
		blsIds = append(blsIds, *bid)
		pubKey := share.GetPublicKey()
		pubKeys = append(pubKeys, *pubKey)
	}

	aggPubKey := bls.PublicKey{}
	err := aggPubKey.Recover(pubKeys, blsIds)

	return &aggPubKey, err
}

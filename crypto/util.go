package crypto

import (
	"encoding/binary"

	"github.com/herumi/bls-eth-go-binary/bls"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

// BLSID turns a uint64 in to a BLS identifier.
// Taken from Attestant's Dirk source code - should be kept in sync with the original.
func BLSID(id uint64) (*bls.ID, error) {
	var res bls.ID
	buf := [8]byte{}
	binary.LittleEndian.PutUint64(buf[:], id)
	if err := res.SetLittleEndian(buf[:]); err != nil {
		return nil, err
	}
	return &res, nil
}

func InitBLS() error {
	return e2types.InitBLS()
}

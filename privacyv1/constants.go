package privacyv1

import "math/big"

const (
	pointCompressed       byte = 0x2
	elGamalCiphertextSize      = 64 // bytes
	schnMultiSigSize           = 65 // bytes
)

const (
	Ed25519KeySize = 32
	AESKeySize = 32
	CommitmentRingSize    = 8
	CommitmentRingSizeExp = 3
	CStringBulletProof = "bulletproof"
)

const (
	MaxSizeInfoCoin = 255  // byte
)

var LInt = new(big.Int).SetBytes([]byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,0x14,0xde,0xf9,0xde,0xa2,0xf7,0x9c,0xd6, 0x58,0x12,0x63,0x1a,0x5c,0xf5,0xd3,0xed})

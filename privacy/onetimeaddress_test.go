package privacy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOneTimeAddress(t *testing.T) {
	for i := 0; i < 10; i++ {
		seed := RandomScalar()
		privSpendKey := GeneratePrivateKey(seed.ToBytesS())
		paymentAddress := GeneratePaymentAddress(privSpendKey)
		viewingKey := GenerateViewingKey(privSpendKey)

		fmt.Printf("Public Spend key: %v\n", paymentAddress.Pk)

		rand := RandomScalar()
		index := 10

		oneTimeAddr, randOTA, err := GenerateOneTimeAddrFromPaymentAddr(paymentAddress, rand, index)
		assert.Equal(t, nil, err)
		assert.Equal(t, Ed25519KeySize, len(oneTimeAddr.ToBytesS()))
		fmt.Printf("oneTimeAddr: %v\n", oneTimeAddr.ToBytesS())

		cmRand := new(Point).ScalarMult(PedCom.G[PedersenPrivateKeyIndex], rand)

		pubSpendKeyFromOneTimeAddr, randOTA2, err := GetPublicKeyFromOneTimeAddress(oneTimeAddr, cmRand, viewingKey.Rk, index)
		fmt.Printf("Public Spend key from one time address: %v\n", pubSpendKeyFromOneTimeAddr.ToBytesS())

		res, randOTA3, err := IsPairOneTimeAddr(oneTimeAddr, cmRand, viewingKey, index)
		assert.Equal(t, true, res)

		assert.Equal(t, randOTA, randOTA2)
		assert.Equal(t, randOTA, randOTA3)
	}
}

func benchmarkGenerateOneTimeAdress(b *testing.B) {
	seed := []byte{1, 2, 3}
	privSpendKey := GeneratePrivateKey(seed)
	paymentAddress := GeneratePaymentAddress(privSpendKey)

	rand := RandomScalar()
	index := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateOneTimeAddrFromPaymentAddr(paymentAddress, rand, index)
	}
}

func benchmarkCheckOneTimeAdress(b *testing.B) {
	seed := []byte{1, 2, 3}
	privSpendKey := GeneratePrivateKey(seed)
	paymentAddress := GeneratePaymentAddress(privSpendKey)
	viewingKey := GenerateViewingKey(privSpendKey)

	rand := RandomScalar()
	cmRand := new(Point).ScalarMult(PedCom.G[0], rand)
	index := 10

	oneTimeAddr, _, _ := GenerateOneTimeAddrFromPaymentAddr(paymentAddress, rand, index)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPairOneTimeAddr(oneTimeAddr, cmRand, viewingKey, index)
	}
}

func BenchmarkGenerateOneTimeAddress(b *testing.B) {
	benchmarkGenerateOneTimeAdress(b)
}

func BenchmarkCheckOneTimeAddress(b *testing.B) {
	benchmarkCheckOneTimeAdress(b)
}

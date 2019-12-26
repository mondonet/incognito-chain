package privacy

// OTA = PubSpendKey * G^(Hash(Hash(PubViewKey^r) || index))
func GenerateOneTimeAddrFromPaymentAddr(paymentAddress PaymentAddress, rand *Scalar, index int) (*Point, *Scalar, error) {
	pubSpendKey, err := new(Point).FromBytesS(paymentAddress.Pk)
	if err != nil {
		return nil, nil, err
	}
	pubViewKey, err := new(Point).FromBytesS(paymentAddress.Tk)
	if err != nil {
		return nil, nil, err
	}

	shareSecretPoint := new(Point).ScalarMult(pubViewKey, rand)
	shareSecretHash := HashToScalar(shareSecretPoint.ToBytesS())
	shareSecretBytes := append(shareSecretHash.ToBytesS(), byte(index))

	randOTA := HashToScalar(shareSecretBytes)

	pubOTA := new(Point).Add(pubSpendKey, new(Point).ScalarMult(PedCom.G[PedersenPrivateKeyIndex], randOTA))

	return pubOTA, randOTA, nil
}

func GetPublicKeyFromOneTimeAddress(oneTimeAddr *Point, cmRand *Point, privViewKey []byte, index int) (*Point, *Scalar, error) {
	privViewKeyPoint := new(Scalar).FromBytesS(privViewKey)

	shareSecretPoint := new(Point).ScalarMult(cmRand, privViewKeyPoint)
	shareSecretHash := HashToScalar(shareSecretPoint.ToBytesS())
	shareSecretBytes := append(shareSecretHash.ToBytesS(), byte(index))

	randOTA := HashToScalar(shareSecretBytes)

	tmp := new(Point).ScalarMult(PedCom.G[PedersenPrivateKeyIndex], randOTA)
	pubSpendKey := new(Point).Sub(oneTimeAddr, tmp)

	return pubSpendKey, randOTA, nil
}

func IsPairOneTimeAddr(oneTimeAddr *Point, cmRand *Point, viewKey ViewingKey, index int) (bool, *Scalar, error) {
	pubSpendKeyFromOneTimeAddr, randOTA, err := GetPublicKeyFromOneTimeAddress(oneTimeAddr, cmRand, viewKey.Rk, index)
	if err != nil {
		return false, randOTA, err
	}

	pubSpendKey, err := new(Point).FromBytesS(viewKey.Pk)
	if err != nil {
		return false, randOTA, err
	}

	isPair := IsPointEqual(pubSpendKeyFromOneTimeAddr, pubSpendKey)
	return isPair, randOTA, nil
}

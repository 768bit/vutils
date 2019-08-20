package Crypto

import "github.com/768bit/vutils/Crypto/Hashing"

type CryptoUtils struct {
	Hashing *Hashing.HashingUtils
}

func NewCryptoUtils() *CryptoUtils {

	return &CryptoUtils{
		Hashing: &Hashing.HashingUtils{},
	}

}

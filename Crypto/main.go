package Crypto

import "gitlab.768bit.com/vann/vutils/Crypto/Hashing"

type CryptoUtils struct {
	Hashing *Hashing.HashingUtils
}

func NewCryptoUtils() *CryptoUtils {

	return &CryptoUtils{
		Hashing: &Hashing.HashingUtils{},
	}

}

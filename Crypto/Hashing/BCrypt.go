package Hashing

import "golang.org/x/crypto/bcrypt"

type bcryptHashingUtils struct {
}

func (*bcryptHashingUtils) HashAndSaltPassword(pwd string) (string, error) {

	bpwd := []byte(pwd)

	hash, err := bcrypt.GenerateFromPassword(bpwd, bcrypt.DefaultCost)

	if err != nil {

		return "", err

	} else {

		return string(hash), nil

	}

}

func (*bcryptHashingUtils) ComparePasswordWithHash(hashedPwd string, pwd string) (bool, error) {

	bpwd := []byte(pwd)

	bhash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(bhash, bpwd)

	if err != nil {

		return false, err

	} else {

		return true, nil

	}

}

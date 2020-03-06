package vutils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
)

type base58Encoder struct{}

//strings.Replace(rpcStr, "-", "", -1)

func (enc base58Encoder) EncodeFromString(u string) (string, error) {
	ba, err := hex.DecodeString(strings.Replace(u, "-", "", -1))
	if err != nil {
		return "", err
	}
	return base58.Encode(ba), nil
}

func (enc base58Encoder) EncodeFromUUID(u *uuid.UUID) (string, error) {
	return enc.EncodeFromString(u.String())
}

func (enc base58Encoder) DecodeToUUID(s string) (uuid.UUID, error) {
	return uuid.FromBytes(base58.Decode(s))
}

func (enc base58Encoder) DecodeToString(s string) (string, error) {
	u, err := uuid.FromBytes(base58.Decode(s))
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

var base58EncoderInst = &base58Encoder{}

type uuidUtils struct {
	uuidStringRx         *regexp.Regexp
	uuidStringNoDashesRx *regexp.Regexp
}

func (uu *uuidUtils) MakeUUID() (*uuid.UUID, error) {

	rpcID, err := uuid.NewRandom()

	if err != nil {

		return nil, err

	}

	return &rpcID, nil

}

func (uu *uuidUtils) MakeUUIDString() (string, error) {

	rpcID, err := uuid.NewRandom()

	if err != nil {

		return "", err

	}

	rpcIDStr := strings.ToLower(rpcID.String())

	return rpcIDStr, nil

}

func (uu *uuidUtils) UUIDToShort(uuidIn string) (string, error) {

	return base58EncoderInst.EncodeFromString(uuidIn)

}

func (uu *uuidUtils) ShortToUUID(shortIn string) (string, error) {

	return base58EncoderInst.DecodeToString(shortIn)

}

func (uu *uuidUtils) MakeUUIDAndString() (*uuid.UUID, string, error) {

	rpcID, err := uuid.NewRandom()

	if err != nil {

		return nil, "", err

	}

	rpcIDStr := strings.ToLower(rpcID.String())

	return &rpcID, rpcIDStr, nil

}

func (uu *uuidUtils) MakeUUIDStringNoDashes() (string, error) {

	rpcUUID, err := uu.MakeUUIDString()

	if err != nil {

		return "", err

	}

	return strings.Replace(rpcUUID, "-", "", -1), nil

}

func (uu *uuidUtils) MakeUUIDAndStringNoDashes() (*uuid.UUID, string, error) {

	rpcUUID, rpcStr, err := uu.MakeUUIDAndString()

	if err != nil {

		return nil, "", err

	}

	return rpcUUID, strings.Replace(rpcStr, "-", "", -1), nil

}

func (uu *uuidUtils) UUIDStringIsValid(u string) bool {

	if !uu.uuidStringRx.MatchString(strings.ToLower(u)) {

		return false

	}

	return true

}

func (uu *uuidUtils) UUIDStringNoDashesIsValid(u string) bool {

	if !uu.uuidStringNoDashesRx.MatchString(strings.ToLower(u)) {

		return false

	}

	return true

}

func (uu *uuidUtils) MakeUUIDFromString(u string) (*uuid.UUID, error) {

	if !uu.UUIDStringIsValid(u) {

		return nil, errors.New(fmt.Sprintf("Supplied UUID String %s is invalid", u))

	}

	rpcID, err := uuid.Parse(u)

	if err != nil {

		return nil, err

	}

	return &rpcID, nil

}

func (uu *uuidUtils) MakeUUIDFromStringNoDashes(u string) (*uuid.UUID, error) {

	if !uu.UUIDStringNoDashesIsValid(u) {

		return nil, errors.New(fmt.Sprintf("Supplied UUID String %s with no dashes is invalid", u))

	}

	//need to add the dashes etc...

	u = strings.ToLower(u[0:7] + "-" + u[8:11] + "-" + u[12:15] + "-" + u[16:19] + "-" + u[20:31])

	rpcID, err := uuid.Parse(u)

	if err != nil {

		return nil, err

	}

	return &rpcID, nil

}

var UUID = uuidUtils{
	uuidStringRx:         regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
	uuidStringNoDashesRx: regexp.MustCompile("^[a-f0-9]{32}$"),
}

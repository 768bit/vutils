package vutils

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

type uuidUtils struct {
	uuidStringRx         *regexp.Regexp
	uuidStringNoDashesRx *regexp.Regexp
}

func (uu *uuidUtils) MakeUUIDString() (string, error) {

	rpcID, err := uuid.NewRandom()

	if err != nil {

		return "", err

	}

	rpcIDStr := strings.ToLower(rpcID.String())

	return rpcIDStr, nil

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

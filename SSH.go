// +build !js

package vutils

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
)

type sshUtils struct {
	DefaultSSHHostKeyCallback ssh.HostKeyCallback
}

func (su *sshUtils) CreateSSHClient(host string, port string, user string, privateKeyPath string, callback ssh.HostKeyCallback) (*ssh.Client, error) {

	privKeyBuffer, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	privKey, err := ssh.ParsePrivateKey(privKeyBuffer)

	if err != nil {
		return nil, err
	}

	pubKey := ssh.PublicKeys(privKey)

	if callback == nil {

		callback = su.DefaultSSHHostKeyCallback

	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			pubKey,
		},
		HostKeyCallback: callback,
	}

	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		return nil, err
	}

	return client, nil

}

func defaultSSHHostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {

	return nil

}

var SSH = &sshUtils{
	DefaultSSHHostKeyCallback: defaultSSHHostKeyCallback,
}

package wg

import (
	"bytes"
	"os/exec"
	"strings"
)

// GeneratePrivateKey 生成私钥
func GeneratePrivateKey() (string, error) {
	cmd := exec.Command("wg", "genkey")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GeneratePublicKey 从私钥生成公钥
func GeneratePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GeneratePresharedKey 生成预共享密钥
func GeneratePresharedKey() (string, error) {
	cmd := exec.Command("wg", "genpsk")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// GenerateKeyPair 生成密钥对
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privateKey, err = GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	publicKey, err = GeneratePublicKey(privateKey)
	if err != nil {
		return "", "", err
	}
	return privateKey, publicKey, nil
}

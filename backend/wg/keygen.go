package wg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GeneratePrivateKey() (string, error) {
	return runKeyCmd("wg", "genkey")
}

func GeneratePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", keyCmdError("wg pubkey", stderr.String(), err)
	}
	return strings.TrimSpace(out.String()), nil
}

func GeneratePresharedKey() (string, error) {
	return runKeyCmd("wg", "genpsk")
}

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

func runKeyCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", keyCmdError(name+" "+strings.Join(args, " "), stderr.String(), err)
	}
	return strings.TrimSpace(out.String()), nil
}

func keyCmdError(cmd, stderr string, err error) error {
	if strings.Contains(err.Error(), "executable file not found") {
		return fmt.Errorf("未找到 wg 命令，请先安装 wireguard-tools")
	}
	msg := strings.TrimSpace(stderr)
	if msg != "" {
		return fmt.Errorf("%s: %s", cmd, msg)
	}
	return fmt.Errorf("%s: %w", cmd, err)
}

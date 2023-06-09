package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func askPassword() []byte {
	fmt.Println("ASK\t| Insert password:")
	bytepassword, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		panic(err)
	}

	return bytepassword
}

func decrypt() string {
	var keystr string

	username := strings.ToLower(getUsername())

	if len(username) < 16 {
		n := 16 - len(username)
		newRune := []rune("agustinlavarello")

		keystr = string(username + string(newRune[0:n]))
	} else {
		newRune := []rune(username)
		keystr = string(newRune)
	}

	key := []byte(keystr)
	var cfg Env = getConfig(getConfigfile())

	if cfg.Password == "" {
		cfg = setPassword(cfg)
	}

	ciphertext, _ := base64.URLEncoding.DecodeString(cfg.Password)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(ciphertext) < aes.BlockSize {
		panic("FATAL | Ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext)
}

func encrypt(b []byte) string {
	var keystr string
	username := strings.ToLower(getUsername())

	if len(username) < 16 {
		n := 16 - len(username)
		newRune := []rune("agustinlavarello")

		keystr = string(username + string(newRune[0:n]))
	} else {
		newRune := []rune(username)
		keystr = string(newRune)
	}

	key := []byte(keystr)
	// plaintext := []byte(s)
	block, err := aes.NewCipher(key)

	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], b)

	return base64.URLEncoding.EncodeToString(ciphertext)
}

func getUsername() string {
	user, _ := user.Current()
	username := user.Username

	if strings.Contains(username, "\\") {
		username = strings.Split(username, "\\")[1]
	}

	return username
}

func getConfigfile() string {
	var configFile string

	if runtime.GOOS == "windows" {
		configFile = os.Getenv("USERPROFILE") + "\\.kube\\occonfig"
	} else {
		configFile = "/home/" + getUsername() + "/.kube/occonfig"
	}

	return configFile
}

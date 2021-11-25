package util

import (
	"encoding/base64"
	"fmt"
	"github.com/Hyperledger-TWGC/tjfoc-gm/sm4"
	"testing"
)

func TestGetUUID(t *testing.T) {
	uuid, err := GetUUID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(uuid), uuid)
}

func TestEncryptAES(t *testing.T) {
	uuid, err := GetUUID()
	if err != nil {
		t.Fatal(err)
	}
	binary := uuid

	fmt.Println(len(binary))
	uuidStr := string(uuid)
	b := []byte(uuidStr)
	fmt.Println(b)
	src := []byte("this is 斯巴达")
	aes, err := EncryptAES(src, binary)
	if err != nil {
		t.Fatal(err)
	}
	decryptAES, err := DecryptAES(aes, binary)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(decryptAES)
}

func TestEncryptSm4(t *testing.T) {
	// 128比特密钥
	key, err := GetUUID()
	if err != nil {
		t.Fatal(err)
	}
	// 128比特iv
	iv := make([]byte, sm4.BlockSize)
	data := []byte("Tongji Fintech Research Institute")
	ciphertxt, err := sm4Encrypt(key, iv, data)
	if err != nil {
		t.Fatal(err)
	}
	s := base64.StdEncoding.EncodeToString(ciphertxt)
	fmt.Println(s)
	d, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatal(d)
	}
	decrypt, err := sm4Decrypt(key, iv, d)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(decrypt))
}

package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
)

const (
	MYKEY = "abcdefgh12345678" //十六字节密匙
	IV    = "aaaabbbb12345678" //CBC模式的初始化向量：与key等长：十六字节
)

func GetUUID() ([]byte, error) {
	return uuid.New().MarshalBinary()
}

func GetUUIDString() (string, error) {
	binary, err := uuid.New().MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(binary), nil
}

func GetSecretKey() ([]byte, error) {
	return GetUUID()
}

// EncryptAES 使用aes进行对称加密
func EncryptAES(src, key []byte) ([]byte, error) {
	//1. 创建并返回一个使用DES算法的cipher.Block接口
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	//2. 对最后一个明文分组进行数据填充
	src = paddingBytes(src, block.BlockSize())
	//3. 创建一个密码分组为链接模式的，底层使用DES加密的BlockMode接口
	cbcDecrypter := cipher.NewCBCEncrypter(block, []byte(IV))
	//4. 加密连续的数据并返回
	dst := make([]byte, len(src))
	cbcDecrypter.CryptBlocks(dst, src)

	return dst, nil
}

// DecryptAES 使用aes进行解密
func DecryptAES(src, key []byte) ([]byte, error) {
	//1. 创建并返回一个使用DES算法的cipher.Block接口
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//2. 创建一个密码分组为链接模式的，底层使用DES解密的BlockMode接口
	cbcDecrypter := cipher.NewCBCDecrypter(block, []byte(IV))
	//3. 数据块解密
	dst := make([]byte, len(src))
	cbcDecrypter.CryptBlocks(dst, src)
	//4. 去掉最后一组填充数据
	newBytes, err := unPaddingBytes(dst)
	if err != nil {
		return nil, err
	}
	return newBytes, nil
}

//pkcs7Padding 填充
func paddingBytes(data []byte, blockSize int) []byte {
	//判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	//补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

//pkcs7UnPadding 填充的反向操作
func unPaddingBytes(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	//获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

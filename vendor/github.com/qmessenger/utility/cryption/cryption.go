package cryption

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
	"crypto/aes"
)

// 平均耗时200us
func EncodeToken(timestamp int64, product, privateKey string) string {
	origData := []byte(strconv.FormatInt(timestamp, 10))
	origData = append(origData, '_')
	origData = append(origData, product...)
	deskey := []byte(Md5(privateKey + product)[:8])
	result, err := DesEncrypt(origData, deskey, deskey)
	if err != nil {
		println("DESEncrypt err:" + err.Error())
		return ""
	}
	return Base64Encode(result)
}

// 平均耗时300us
func DecodeToken(token, product, privateKey string) string {
	ciphertext, err := Base64Decode(token)
	if err != nil {
		println("base64 decode err:" + err.Error())
		return ""
	}

	deskey := []byte(Md5(privateKey + product)[:8])
	result, err := DesDecrypt(ciphertext, deskey, deskey)
	if err != nil {
		println("DesDecrypt err:" + err.Error())
	}

	timeProduct := strings.SplitN(string(result), "_", 2)
	if len(timeProduct) == 1 {
		// 旧的检测方式，解密的数据只有时间戳
	} else if len(timeProduct) == 2 && timeProduct[1] != product {
		// 新的检测方式，解密的数据为时间戳+"_"+产品名
		fmt.Println("[checkToken] product:" + product + "|err:bad product|data:" + timeProduct[1])
		return ""
	}

	sendtime, err := strconv.ParseInt(timeProduct[0], 10, 64)
	if err != nil {
		fmt.Println("[checkToken] product:" + product + "|err:" + err.Error())
		return ""
	}
	// 与当前时间比较
	now := time.Now().Unix()
	if now-sendtime > 10 || now-sendtime < 0 {
		fmt.Println("[checkToken] product:" + product + "|err: invalid time scope CenterTIme:" + strconv.FormatInt(now, 10) + "  SendTime:" + timeProduct[0])
		return ""
	}
	return string(result)
}

func RsaEncrypt(origData, publickKey []byte) ([]byte, error) {
	block, _ := pem.Decode(publickKey)
	if block == nil {
		return nil, errors.New("public key error!")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func RsaDecrypt(ciphertext, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}

func DesEncrypt(origData, desKey, iv []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	if len(iv) < 8 {
		return nil, errors.New("The invalid length of iv: " + strconv.Itoa(len(iv)) + ", need 8")
	}
	block, err := des.NewCipher(desKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:8])
	srcLen := len(origData)
	if srcLen%blockSize != 0 {
		return nil, errors.New("The length of origData must be a multiple of the block size")
	}
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func DesDecrypt(ciphertext, desKey, iv []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	if len(iv) < 8 {
		return nil, errors.New("The invalid length of iv: " + strconv.Itoa(len(iv)) + ", need 8")
	}
	block, err := des.NewCipher(desKey)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv[:8])
	cipherLen := len(ciphertext)
	// 避免因为go的bug而panic
	if cipherLen%block.BlockSize() != 0 {
		return nil, errors.New("The length of ciphertext must be a multiple of the block size")
	}
	origData := make([]byte, cipherLen)
	blockMode.CryptBlocks(origData, ciphertext)
	if len(origData) < 1 {
		return nil, errors.New("invalid origData length")
	}
	return PKCS5UnPadding(origData), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(ciphertext []byte) []byte {
	unpadding := int(ciphertext[len(ciphertext)-1])
	return ciphertext[:(len(ciphertext) - unpadding)]
}

func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func Base64Decode(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}

func Md5(text string) string {
	hashMd5 := md5.New()
	io.WriteString(hashMd5, text)
	return fmt.Sprintf("%x", hashMd5.Sum(nil))
}

func Urldecode(text string) (string, error) {
	m, err := url.ParseQuery("field=" + text)
	if err != nil {
		return "", err
	}
	return m["field"][0], nil
}

func Rc4Encrypt(originData []byte, key []byte) ([]byte, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(originData))
	c.XORKeyStream(ciphertext, originData)

	return ciphertext, nil
}

func Rc4Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	originData := make([]byte, len(ciphertext))
	c.XORKeyStream(originData, ciphertext)

	return originData, nil
}


// AES加密，经济系统使用了ECB而非CBC的方式
func AesEncrypt(origData, aesKey []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	if len(aesKey) < 16 {
		return nil, errors.New("The invalid length of aes key: " + strconv.Itoa(len(aesKey)) + ", need 16")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := NewECBEncrypter(block)
	srcLen := len(origData)
	if srcLen%blockSize != 0 {
		return nil, errors.New("The length of origData must be a multiple of the block size")
	}
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AES解密，经济系统使用了ECB而非CBC的方式
func AesDecrypt(ciphertext, aesKey []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	if len(aesKey) < 16 {
		return nil, errors.New("The invalid length of aes key: " + strconv.Itoa(len(aesKey)) + ", need 16")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	blockMode := NewECBDecrypter(block)
	cipherLen := len(ciphertext)
	// 避免因为go的bug而panic
	if cipherLen%block.BlockSize() != 0 {
		return nil, errors.New("The length of ciphertext must be a multiple of the block size")
	}
	origData := make([]byte, cipherLen)
	blockMode.CryptBlocks(origData, ciphertext)
	if len(origData) < 1 {
		return nil, errors.New("invalid origData length")
	}
	return PKCS5UnPadding(origData), nil
}

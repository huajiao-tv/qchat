package cryption

import (
	"bytes"
	"testing"
)

var (
	pubKey = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC79aaj3GEdtzUM1hsa9gtV6wbd
iZr0BApUzNpDyWgycMDcY9Gf48Kl4MRG2qi/oWrcxmHM4p9KXyHqMEzXjdJVYX3b
v6hVX2r8CakHTHaS3AAXPEdaLfDundI8Ru/YiJ7MI0CfWCmsYCAJvno6110qWqnk
JYRcVaA4oEhzPqz0cwIDAQAB
-----END PUBLIC KEY-----
    `
	privKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC79aaj3GEdtzUM1hsa9gtV6wbdiZr0BApUzNpDyWgycMDcY9Gf
48Kl4MRG2qi/oWrcxmHM4p9KXyHqMEzXjdJVYX3bv6hVX2r8CakHTHaS3AAXPEda
LfDundI8Ru/YiJ7MI0CfWCmsYCAJvno6110qWqnkJYRcVaA4oEhzPqz0cwIDAQAB
AoGBALf9go8alnJ5OdQD7mqY+YW0WHcaUXWWUuqp0OrUSEw/9XqHt9a1JIAuItRd
DRzxDONqyqe+G0G5GEDf4QiMSpweKWS4vsJNLiK1bEqTnOiD9Ig7blpVNDG7RrXX
BOyCEq5Bw1VEXA2X9xR0SkMWZEqFSIND99pLlXoko3zf0TTZAkEA3JxHmPDzi7tu
hgoS1p+eOpgzIpwWrRVke4lf7m3cEXCNs78CeHyKAFXYLewn/X7/k/Ot8h8GirXr
rFsQb9afNwJBANocgqL1GCvVUaapQRn6MuXvDh2wZXfKBY+cB4kawkTIzA4kBTm6
TZAXdpUKmC19i+LBr6P5Z3Lc1n7dV58ZWqUCQFuZ2HC8u6Ntc/rb++554G1b/P+F
6DR+CXbyF48ctp/XKD9WNGRq8bIp8tU+lWxAa0a3i6ZZE5JM70pllXGaoAkCQGLH
wL5uxCit7tHNG8fZEY4jS0BU8E9lNkmI/7yvWsZuLkRFOfygDJqylakAaFVJ472p
vJNF0/0oWRiRxCoxAGUCQQCpjEBHMU4iZH9TJRkouf2CcKGr3C6RwnF+I9SMT+kM
O6J0lu0wLY6/HV1L+1RjPfqeh4Wz6nqU+Dprm7BiVRLu
-----END RSA PRIVATE KEY-----
    `
)

type caseData struct {
	origData   []byte
	ciphertext []byte
}

var RSAtestCases = []caseData{
	{
		[]byte("polaris@baohe"),
		[]byte{},
	},
}

// 加密之后马上解密
func TestRsaEncrypt(t *testing.T) {
	for _, tc := range RSAtestCases {
		ciphertext, err := RsaEncrypt(tc.origData, []byte(pubKey))
		if err != nil {
			t.Error(err)
		}
		origData, err := RsaDecrypt(ciphertext, []byte(privKey))
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(origData, tc.origData) {
			t.Errorf("after rsa encrypt/decrypt: %v; should be: %v", string(origData), string(tc.origData))
		}
	}
}

var Base64testCases = []caseData{
	{
		[]byte("polaris"),
		[]byte("cG9sYXJpcw=="),
	},
}

func TestBase64Encode(t *testing.T) {
	for _, tc := range Base64testCases {
		encoded := Base64Encode(tc.origData)
		if !bytes.Equal([]byte(encoded), tc.ciphertext) {
			t.Errorf("after base64 encode: %v; should be: %s", encoded, tc.ciphertext)
		}
	}
}

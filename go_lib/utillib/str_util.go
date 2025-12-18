package utillib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"github.com/forgoer/openssl"
	"github.com/google/uuid"
	"strings"
)

func StrGzip(input string) (string, error) {
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	_, err := gw.Write([]byte(input))
	if err != nil {
		return "", err
	}
	if err := gw.Close(); err != nil {
		return "", err
	}
	// 将压缩后的数据进行 Base64 编码
	encoded := base64.StdEncoding.EncodeToString(b.Bytes())
	return encoded, nil
}

func DeStrGzip(encoded string) (string, error) {
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	gr, err := gzip.NewReader(bytes.NewBuffer(compressed))
	if err != nil {
		return "", err
	}
	defer gr.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(gr)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func GetUUid() string {
	uuidString := strings.ReplaceAll(uuid.New().String(), "-", "")
	return uuidString
}

func AesCBCEncryptStr(src, aesIv, aesKey string) string {
	dst, err := openssl.AesCBCEncrypt([]byte(src), []byte(aesKey),
		[]byte(aesIv), openssl.PKCS5_PADDING)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(dst)
}

func AesCBCDecryptStr(src, aesIv, aesKey string) string {
	srcBytes, _ := base64.StdEncoding.DecodeString(src)
	srcDecode, err := openssl.AesCBCDecrypt(srcBytes, []byte(aesKey),
		[]byte(aesIv), openssl.PKCS5_PADDING)
	if err != nil {
		return ""
	}
	return string(srcDecode)
}

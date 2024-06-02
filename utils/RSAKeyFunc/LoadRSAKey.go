package RSAKeyFunc

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func LoadPrivateKeyFromFile(filename string) (*rsa.PrivateKey, error) {
	// Открытие файла с приватным ключом
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Чтение данных из файла
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	keyBytes := make([]byte, fileSize)
	_, err = file.Read(keyBytes)
	if err != nil {
		return nil, err
	}

	// Декодирование приватного ключа
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

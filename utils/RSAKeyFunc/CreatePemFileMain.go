package RSAKeyFunc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func savePrivateKeyToFile(privateKey *rsa.PrivateKey, filename string) error {
	// Преобразование приватного ключа в формат PKCS#8
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// Создание файла и запись приватного ключа в него
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, privateKeyPEM)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	err := savePrivateKeyToFile(privateKey, "..\\..\\configs\\private_key.pem")
	if err != nil {
		panic(err)
	}
}

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

var (
	encryptionKey []byte
	hmacKey       []byte
)

// InitCrypto valida y carga ENCRYPTION_KEY y HMAC_KEY desde el entorno.
// Debe llamarse en startup; falla con log.Fatal si las claves son inválidas.
func InitCrypto() error {
	encKeyHex := os.Getenv("ENCRYPTION_KEY")
	if len(encKeyHex) != 64 {
		return fmt.Errorf("ENCRYPTION_KEY debe ser un hex de 32 bytes (64 caracteres), tiene %d", len(encKeyHex))
	}
	decoded, err := hex.DecodeString(encKeyHex)
	if err != nil {
		return fmt.Errorf("ENCRYPTION_KEY no es hex válido: %w", err)
	}
	encryptionKey = decoded

	hmacKeyHex := os.Getenv("HMAC_KEY")
	if len(hmacKeyHex) != 64 {
		return fmt.Errorf("HMAC_KEY debe ser un hex de 32 bytes (64 caracteres), tiene %d", len(hmacKeyHex))
	}
	hmacDecoded, err := hex.DecodeString(hmacKeyHex)
	if err != nil {
		return fmt.Errorf("HMAC_KEY no es hex válido: %w", err)
	}
	hmacKey = hmacDecoded

	return nil
}

// Encrypt cifra plaintext con AES-256-GCM usando un nonce aleatorio.
// Retorna base64(nonce || ciphertext || tag).
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", fmt.Errorf("crypto no inicializado: llamar a InitCrypto primero")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("error creando cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creando GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("error generando nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt descifra un valor producido por Encrypt.
func Decrypt(encoded string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", fmt.Errorf("crypto no inicializado: llamar a InitCrypto primero")
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 inválido: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("error creando cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creando GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext demasiado corto")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error desencriptando: %w", err)
	}

	return string(plaintext), nil
}

// HMACField retorna HMAC-SHA256 hex del valor normalizado.
// Se usa para columnas de búsqueda exacta sobre datos encriptados.
func HMACField(value string) string {
	mac := hmac.New(sha256.New, hmacKey)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

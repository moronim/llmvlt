package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	saltLen  = 16
	nonceLen = 12
	keyLen   = 32 // AES-256

	// Argon2id parameters
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
)

// secretEntry stores a value along with metadata.
type secretEntry struct {
	Value   string    `json:"value"`
	SetAt   time.Time `json:"set_at"`
	Version int       `json:"version"`
}

// vaultData is the serialized vault structure.
type vaultData struct {
	Secrets        map[string]secretEntry `json:"secrets"`
	ActiveProvider string                 `json:"active_provider,omitempty"`
}

// Vault is the in-memory representation of the encrypted store.
type Vault struct {
	data vaultData
}

// NewVault creates a new empty vault.
func NewVault() *Vault {
	return &Vault{
		data: vaultData{
			Secrets: make(map[string]secretEntry),
		},
	}
}

// Set stores a secret value. Increments version if key already exists.
func (v *Vault) Set(key, value string) {
	existing, ok := v.data.Secrets[key]
	version := 1
	if ok {
		version = existing.Version + 1
	}
	v.data.Secrets[key] = secretEntry{
		Value:   value,
		SetAt:   time.Now(),
		Version: version,
	}
}

// Get retrieves a secret value. Returns ("", false) if not found.
func (v *Vault) Get(key string) (string, bool) {
	e, ok := v.data.Secrets[key]
	if !ok {
		return "", false
	}
	return e.Value, true
}

// Keys returns all secret key names.
func (v *Vault) Keys() []string {
	keys := make([]string, 0, len(v.data.Secrets))
	for k := range v.data.Secrets {
		keys = append(keys, k)
	}
	return keys
}

// All returns all secrets as key-value pairs. Respects active provider filter.
func (v *Vault) All() map[string]string {
	out := make(map[string]string, len(v.data.Secrets))
	for k, e := range v.data.Secrets {
		out[k] = e.Value
	}
	return out
}

// Unset removes a secret.
func (v *Vault) Unset(key string) {
	delete(v.data.Secrets, key)
}

// SecretAgeDays returns how many days since a secret was last set.
func (v *Vault) SecretAgeDays(key string) int {
	e, ok := v.data.Secrets[key]
	if !ok {
		return 0
	}
	return int(time.Since(e.SetAt).Hours() / 24)
}

// SetActiveProvider sets the provider filter for injection.
func (v *Vault) SetActiveProvider(provider string) {
	if provider == "all" {
		v.data.ActiveProvider = ""
	} else {
		v.data.ActiveProvider = provider
	}
}

// GetActiveProvider returns the current active provider, or "" for all.
func (v *Vault) GetActiveProvider() string {
	return v.data.ActiveProvider
}

// --- Encryption / Persistence ---

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keyLen)
}

func encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("could not generate salt: %w", err)
	}

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create GCM: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("could not generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Format: salt || nonce || ciphertext
	out := make([]byte, 0, saltLen+nonceLen+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

func decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < saltLen+nonceLen+1 {
		return nil, fmt.Errorf("encrypted data is too short")
	}

	salt := data[:saltLen]
	nonce := data[saltLen : saltLen+nonceLen]
	ciphertext := data[saltLen+nonceLen:]

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("could not create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("could not create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed — wrong password or corrupted vault")
	}

	return plaintext, nil
}

// Save encrypts and writes the vault to disk.
func Save(path, password string, v *Vault) error {
	plaintext, err := json.Marshal(v.data)
	if err != nil {
		return fmt.Errorf("could not serialize vault: %w", err)
	}

	encrypted, err := encrypt(plaintext, password)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("could not write vault file: %w", err)
	}

	return nil
}

// Load reads and decrypts the vault from disk.
func Load(path, password string) (*Vault, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read vault file: %w", err)
	}

	plaintext, err := decrypt(data, password)
	if err != nil {
		return nil, err
	}

	var vd vaultData
	if err := json.Unmarshal(plaintext, &vd); err != nil {
		return nil, fmt.Errorf("could not parse vault data: %w", err)
	}

	if vd.Secrets == nil {
		vd.Secrets = make(map[string]secretEntry)
	}

	return &Vault{data: vd}, nil
}

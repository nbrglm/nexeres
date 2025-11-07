// Package password provides utilities for hashing and verifying passwords.
//
// The hashing algorithm can be set in the configuration file.
// The default algorithm is bcrypt, but others like argon2id can be used as well.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/nbrglm/nexeres/config"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the provided password using the configured algorithm.
func HashPassword(password string) (string, error) {
	if config.C.Password.Algorithm == config.BcryptPasswordHashingAlgorithm {
		return hashPasswordBCrypt(password)
	} else if config.C.Password.Algorithm == config.Argon2idPasswordHashingAlgorithm {
		return hashPasswordArgon2id(password)
	}

	return password, nil // Default case, return the password as is
}

func hashPasswordBCrypt(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), config.C.Password.Bcrypt.Cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func hashPasswordArgon2id(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, config.C.Password.Argon2id.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, uint32(config.C.Password.Argon2id.Iterations), uint32(config.C.Password.Argon2id.Memory), uint8(config.C.Password.Argon2id.Parallelism), uint32(config.C.Password.Argon2id.KeyLength))

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, config.C.Password.Argon2id.Memory, config.C.Password.Argon2id.Iterations, config.C.Password.Argon2id.Parallelism, b64Salt, b64Hash)
	return encodedHash, nil
}

// VerifyPasswordMatch checks if the hashed password matches the provided password.
func VerifyPasswordMatch(hashedPassword, password string) bool {
	if strings.HasPrefix(hashedPassword, "$argon2id$") {
		return verifyPasswordArgon2id(hashedPassword, password)
	} else if strings.HasPrefix(hashedPassword, "$2a$") || strings.HasPrefix(hashedPassword, "$2b$") || strings.HasPrefix(hashedPassword, "$2y$") {
		return verifyPasswordBCrypt(hashedPassword, password)
	}
	if config.C.Password.Algorithm == config.BcryptPasswordHashingAlgorithm {

	} else if config.C.Password.Algorithm == config.Argon2idPasswordHashingAlgorithm {
		// Implement Argon2id verification here if needed
	}

	return hashedPassword == password
}

func verifyPasswordBCrypt(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func verifyPasswordArgon2id(hashedPassword, password string) bool {
	params, salt, hash, err := decodeArgon2idHash(hashedPassword)
	if err != nil {
		return false
	}

	// Generate the hash for the provided password using the same parameters.
	computedHash := argon2.IDKey([]byte(password), salt, uint32(params.iterations), uint32(params.memory), uint8(params.parallelism), params.keyLength)

	return subtle.ConstantTimeCompare(computedHash, hash) == 1
}

type Argon2idParams struct {
	memory      int
	iterations  int
	parallelism int
	saltLength  uint32
	keyLength   uint32
}

func decodeArgon2idHash(encodedHash string) (p *Argon2idParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid Argon2id hash format: %s", encodedHash)
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("unsupported Argon2 version: %d", version)
	}

	p = &Argon2idParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

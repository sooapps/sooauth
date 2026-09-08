package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash   = errors.New("invalid password hash")
	ErrIncompatible  = errors.New("incompatible argon2 version")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

func Hash(plain string) (string, error) {
	if len(plain) < 8 {
		return "", ErrWeakPassword
	}

	salt := make([]byte, DefaultParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(plain),
		salt,
		DefaultParams.Iterations,
		DefaultParams.Memory,
		DefaultParams.Parallelism,
		DefaultParams.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultParams.Memory,
		DefaultParams.Iterations,
		DefaultParams.Parallelism,
		b64Salt,
		b64Hash,
	), nil
}

func Verify(plain, encoded string) (bool, error) {
	params, salt, hash, err := decode(encoded)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey(
		[]byte(plain),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return subtle.ConstantTimeCompare(hash, other) == 1, nil
}

func decode(encoded string) (Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Params{}, nil, nil, ErrInvalidHash
	}

	versionPart := strings.TrimPrefix(parts[2], "v=")
	version, err := strconv.Atoi(versionPart)
	if err != nil || version != argon2.Version {
		return Params{}, nil, nil, ErrIncompatible
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Params{}, nil, nil, ErrInvalidHash
	}

	return Params{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		KeyLength:     uint32(len(hash)),
	}, salt, hash, nil
}

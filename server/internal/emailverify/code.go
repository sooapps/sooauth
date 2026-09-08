package emailverify

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"

	"github.com/sooapps/sooauth/server/internal/crypto/token"
)

var digitsOnly = regexp.MustCompile(`^\d{6}$`)

func GenerateCode() (plain string, hash string, err error) {
	var n uint32
	if err := binary.Read(rand.Reader, binary.BigEndian, &n); err != nil {
		return "", "", err
	}
	plain = fmt.Sprintf("%06d", n%1_000_000)
	return plain, token.Hash(plain), nil
}

func NormalizeCode(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "-", "")
	return raw
}

func ValidCode(code string) bool {
	return digitsOnly.MatchString(code)
}

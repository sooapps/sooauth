package passwordpolicy

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrWeak = errors.New("password does not meet policy")

type Policy struct {
	MinLength        int
	RequireUppercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

func Default() Policy {
	return Policy{MinLength: 8}
}

func (p Policy) Normalize() Policy {
	if p.MinLength < 8 {
		p.MinLength = 8
	}
	if p.MinLength > 128 {
		p.MinLength = 128
	}
	return p
}

func (p Policy) Validate(plain string) error {
	p = p.Normalize()
	if len(plain) < p.MinLength {
		return ErrWeak
	}
	if p.RequireUppercase {
		has := false
		for _, r := range plain {
			if unicode.IsUpper(r) {
				has = true
				break
			}
		}
		if !has {
			return ErrWeak
		}
	}
	if p.RequireNumber {
		has := false
		for _, r := range plain {
			if unicode.IsDigit(r) {
				has = true
				break
			}
		}
		if !has {
			return ErrWeak
		}
	}
	if p.RequireSpecial {
		has := false
		for _, r := range plain {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				has = true
				break
			}
		}
		if !has {
			return ErrWeak
		}
	}
	return nil
}

func (p Policy) Describe() string {
	p = p.Normalize()
	var parts []string
	parts = append(parts, "at least "+strconv.Itoa(p.MinLength)+" characters")
	if p.RequireUppercase {
		parts = append(parts, "one uppercase letter")
	}
	if p.RequireNumber {
		parts = append(parts, "one number")
	}
	if p.RequireSpecial {
		parts = append(parts, "one special character")
	}
	return strings.Join(parts, ", ")
}

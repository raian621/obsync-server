package database

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("password does not match passhash")
var ErrArgon2idVersion = errors.New("argon2id version does not match the version used in hash")

const ISO_8601_FORMAT = "2006-01-02 15:04:05.999999999+00:00"

type Scannable interface {
	Scan(dest ...any) error
}

type Argon2idParams struct {
	hash        []byte
	salt        []byte
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func (p *Argon2idParams) Parse(argonStr string) error {
	sections := strings.Split(argonStr, "$")
	var version int

	_, err := fmt.Sscanf(sections[2], "v=%d", &version)
	if err != nil {
		return err
	}
	_, err = fmt.Sscanf(
		sections[3],
		"m=%d,t=%d,p=%d",
		&p.memory,
		&p.iterations,
		&p.parallelism,
	)
	if err != nil {
		return err
	}
	saltStr := sections[4]
	hashStr := sections[5]
	if argon2.Version != version {
		return ErrArgon2idVersion
	}
	p.keyLength = uint32(base64.RawStdEncoding.DecodedLen(len(hashStr)))
	p.saltLength = uint32(base64.RawStdEncoding.DecodedLen(len(saltStr)))
	p.hash = make([]byte, p.keyLength)
	_, err = base64.RawStdEncoding.Decode(p.hash, []byte(hashStr))
	if err != nil {
		return err
	}
	p.salt = make([]byte, p.saltLength)
	_, err = base64.RawStdEncoding.Decode(p.salt, []byte(saltStr))
	if err != nil {
		return err
	}

	return err
}

func (p *Argon2idParams) ToString() string {
	b64hash := base64.RawStdEncoding.EncodeToString(p.hash)
	b64salt := base64.RawStdEncoding.EncodeToString(p.salt)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64salt,
		b64hash,
	)
}

func HashPassword(password string) (string, error) {
	return argon2idHash(password, &Argon2idParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	})
}

func ValidateHash(password, passhash string) error {
	params := &Argon2idParams{}
	if err := params.Parse(passhash); err != nil {
		return err
	}
	params.hash = []byte{}
	otherPasshash, err := argon2idHash(password, params)
	if err != nil {
		return err
	}
	if otherPasshash != passhash {
		return ErrInvalidPassword
	}
	return nil
}

func argon2idHash(password string, params *Argon2idParams) (string, error) {
	var err error
	if len(params.salt) == 0 {
		params.salt, err = randomBytes(uint(params.saltLength))
		if err != nil {
			return "", err
		}
	}

	params.hash = argon2.IDKey(
		[]byte(password),
		params.salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	return params.ToString(), nil
}

func randomBytes(length uint) ([]byte, error) {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func generateKey(length uint) (string, error) {
	if b, err := randomBytes(length); err != nil {
		return "", err
	} else {
		return base64.RawStdEncoding.EncodeToString(b), nil
	}
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

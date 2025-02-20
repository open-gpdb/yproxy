package crypt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/yezzey-gp/yproxy/config"
	"github.com/yezzey-gp/yproxy/pkg/ylogger"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type KeyVersion int

const (
	SingleKeyUsed = KeyVersion(iota + 1) // Single key is used for encryption
	KEKDEKUsed
)

type Crypter interface {
	Decrypt(reader io.ReadCloser, keyVersion KeyVersion) (io.Reader, error)
	Encrypt(writer io.WriteCloser) (io.WriteCloser, KeyVersion, error)
}

type GPGCrypter struct {
	EntityList       openpgp.EntityList
	KEKDEKEntityList openpgp.EntityList

	cnf *config.Crypto
}

func NewCrypto(cnf *config.Crypto) (Crypter, error) {
	cr := &GPGCrypter{
		cnf: cnf,
	}

	err := cr.loadSecret()
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func (g *GPGCrypter) readKey(path string) (io.Reader, error) {
	byteData, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return bytes.NewReader(byteData), nil
}

func (g *GPGCrypter) readPGPKey() (openpgp.EntityList, error) {
	gpgKeyReader, err := g.readKey(g.cnf.GPGKeyPath)

	if err != nil {
		return nil, err
	}

	entityList, err := openpgp.ReadArmoredKeyRing(gpgKeyReader)

	if err != nil {
		return nil, err
	}

	return entityList, nil
}

func (g *GPGCrypter) readGPGKEKDEK() (openpgp.EntityList, error) {
	gpgKEKReader, err := g.readKey(g.cnf.GPGKEKPath)

	if err != nil {
		return nil, err
	}

	kEKEntityList, err := openpgp.ReadArmoredKeyRing(gpgKEKReader)

	if err != nil {
		return nil, err
	}

	dEKData, err := g.readKey(g.cnf.GPGDEKPath)

	if err != nil {
		return nil, err
	}

	md, err := openpgp.ReadMessage(dEKData, kEKEntityList, nil, nil)
	if err != nil {
		return nil, err
	}

	return openpgp.ReadArmoredKeyRing(md.UnverifiedBody)
}

func (g *GPGCrypter) loadSecret() error {
	success := false
	errs := make([]error, 0)
	entityList, err := g.readPGPKey()

	if err != nil {
		errs = append(errs, err)
	} else {
		success = true
		g.EntityList = entityList
	}

	kEKEntityList, err := g.readGPGKEKDEK()
	if err != nil {
		errs = append(errs, err)
	} else {
		success = true
		g.KEKDEKEntityList = kEKEntityList
	}

	if !success {
		msgs := make([]string, len(errs))
		for i, err := range errs {
			msgs[i] = errors.WithStack(err).Error()
		}
		return fmt.Errorf(strings.Join(msgs, "\n"))
	}
	return nil
}

func (g *GPGCrypter) Decrypt(reader io.ReadCloser, keyVersion KeyVersion) (io.Reader, error) {
	ylogger.Zero.Debug().Str("gpg path", g.cnf.GPGKeyPath).Msg("loaded gpg key")

	var entityList openpgp.EntityList
	switch keyVersion {
	case SingleKeyUsed:
		entityList = g.EntityList
	case KEKDEKUsed:
		entityList = g.KEKDEKEntityList
	default:
		return nil, fmt.Errorf("incorrect key version %d", int(keyVersion))
	}

	md, err := openpgp.ReadMessage(reader, entityList, nil, nil)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return md.UnverifiedBody, nil
}

func (g *GPGCrypter) Encrypt(writer io.WriteCloser) (io.WriteCloser, KeyVersion, error) {
	var entityList openpgp.EntityList
	keyVersion := SingleKeyUsed
	if g.KEKDEKEntityList != nil {
		entityList = g.KEKDEKEntityList
		keyVersion = KEKDEKUsed
		ylogger.Zero.Debug().
			Str("KEK path", g.cnf.GPGDEKPath).
			Str("DEK path", g.cnf.GPGDEKPath).
			Msg("loaded gpg KEK & DEK")
	} else {
		entityList = g.EntityList
		ylogger.Zero.Debug().Str("gpg path", g.cnf.GPGKeyPath).Msg("loaded gpg key")
	}
	encryptedWriter, err := openpgp.Encrypt(writer, entityList, nil, nil, nil)

	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	return encryptedWriter, keyVersion, nil
}

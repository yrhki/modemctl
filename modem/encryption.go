package modem

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

const (
	rsaModulus   = "BEB90F8AF5D8A7C7DA8CA74AC43E1EE8A48E6860C0D46A5D690BEA082E3A74E1571F2C58E94EE339862A49A811A31BB4A48F41B3BCDFD054C3443BB610B5418B3CBAFAE7936E1BE2AFD2E0DF865A6E59C2B8DF1E8D5702567D0A9650CB07A43DE39020969DF0997FCA587D9A8AE4627CF18477EC06765DF3AA8FB459DD4C9AF3"
	rsaExponent  = "10001"

	// https://www.elektroda.com/rtvforum/topic3438655.html
	confKey      = "3E4F5612EF64305955D543B0AE350880"
	confIV       = "8049E91025A6B54876C3B4868090D3FC"
	passKey      = "3E4F5612EF64305955D543B0AE3508807968905960C44D37"
	passIV       = "8049E91025A6B548"
)

var (
	cipherConf cipher.BlockMode
	cipherPass cipher.BlockMode
	keyPublic  *rsa.PublicKey
)

func init() {
	// cipherConf
	keyByte, err := hex.DecodeString(confKey)
	if err != nil { panic(err) }

	block, err := aes.NewCipher(keyByte)
	if err != nil { panic(err) }

	keyByte, err = hex.DecodeString(confIV)
	if err != nil { panic(err) }

	cipherConf = cipher.NewCBCDecrypter(block, keyByte)

	// cipherPass
	keyByte, err = hex.DecodeString(passKey)
	if err != nil { panic(err) }

	block, err = des.NewTripleDESCipher(keyByte)
	if err != nil { panic(err) }

	keyByte, err = hex.DecodeString(passIV)
	if err != nil { panic(err) }

	cipherPass = cipher.NewCBCDecrypter(block, keyByte)

	// keyPublic
	n := new(big.Int)
	_, ok := n.SetString(rsaModulus, 16)
	if !ok { panic("") }

	e, err := strconv.ParseInt(rsaExponent, 16, 0)
	if err != nil { panic(err) }

	keyPublic = &rsa.PublicKey{
		N: n,
		E: int(e),
	}
}

func decrypt(enc *bytes.Buffer, decrypter cipher.BlockMode) *bytes.Buffer {
	dec := make([]byte, enc.Len())
	decrypter.CryptBlocks(dec, enc.Bytes())

	return bytes.NewBuffer(dec)
}


func encryptPassword(username, password string, t *token) (string, error) {
	hash := sha256.Sum256([]byte(password))
	hashS := []byte(fmt.Sprintf("%x", hash))
	b64 := base64.StdEncoding.EncodeToString(hashS)

	passHash := sha256.Sum256([]byte(username + b64 + t.csrfParam + t.csrfToken))
	passHashString := []byte(fmt.Sprintf("%x", passHash))

	passEnc, err := rsa.EncryptPKCS1v15(rand.Reader, keyPublic, passHashString)
	if err != nil { return "", err }

	return base64.StdEncoding.EncodeToString(passEnc), nil
}


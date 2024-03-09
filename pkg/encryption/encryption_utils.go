package encryptionutils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	cidUtil "github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"golang.org/x/crypto/pbkdf2"
)

func GetAESKeySaltIV(personalSignature string, salt, iv []uint8) (aesKey []uint8, saltThis []uint8, ivThis []uint8, err error) {
	//get passwordBytes from signature
	passwordBytes := []uint8(personalSignature)
	//get PBKDF2 key
	aesKey, saltThis, ivThis, err = GetPBKDF2Key(passwordBytes, salt, iv)
	// get aesKey, salt and iv
	return aesKey, saltThis, ivThis, err
}

func GetPBKDF2Key(passwordBytes []uint8, salt []uint8, iv []uint8) (aesKey []uint8, saltThis, ivThis []uint8, err error) {

	//get aesKey
	aesKey, err = GetAESKey(passwordBytes, salt, iv)

	return aesKey, salt, iv, err
}

func GetAESKey(passwordBytes []uint8, salt []uint8, iv []uint8) (aesKey []uint8, err error) {
	//get aesKey
	pbkdf2Key := pbkdf2.Key(passwordBytes, salt, 250000, 32, sha256.New)

	//get aesKey
	aesKey = pbkdf2Key[:32]

	return
}

func GenerateRandomBytes(length int) (bytes []uint8, err error) {
	bytes = make([]uint8, length)
	_, err = rand.Read(bytes)
	return
}

func EncryptBuffer(personalSignature string, content []uint8) (encryptedBytes []uint8, err error) {
	// get aesKey, salt and iv

	//get salt
	//if salt is not provided, generate a random salt
	var salt []uint8
	var iv []uint8
	if salt == nil {
		//generate empty salt of length 16 and filled with 0
		salt = make([]uint8, 16)
		for i := range salt {
			salt[i] = 0
		}
	} else {
		salt = []uint8(salt)
	}
	if iv == nil {
		//generate empty iv of length 12 and filled with 0
		iv = make([]uint8, 12)
		for i := range iv {
			iv[i] = 0
		}
	} else {
		iv = []uint8(iv)
	}

	aesKey, salt, iv, err := GetAESKeySaltIV(personalSignature, salt, iv)
	if err != nil {
		return nil, err
	}

	//encrypt file metadata
	cidOriginalEncryptedBuffer, err := EncryptBytes(content, aesKey, salt, iv)
	if err != nil {
		return nil, err
	}

	return cidOriginalEncryptedBuffer, nil

}

func EncryptBytes(content []uint8, aesKey []uint8, salt, iv []uint8) (encryptedBytes []uint8, err error) {
	//encrypt content
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	encryptedBytes = aesgcm.Seal(nil, iv, content, nil)

	//set salt to the beginning of encryptedBytes and iv to the beginning of salt
	encryptedBytes = getResultBytes(encryptedBytes, salt, iv)

	return encryptedBytes, nil
}

func getResultBytes(cipherBytesArray []uint8, salt []uint8, iv []uint8) []uint8 {
	resultBytes := make([]uint8, len(cipherBytesArray)+len(salt)+len(iv))

	copy(resultBytes[0:len(salt)], salt)
	copy(resultBytes[len(salt):len(salt)+len(iv)], iv)
	copy(resultBytes[len(salt)+len(iv):], cipherBytesArray)

	return resultBytes
}

func BufferToBase64Url(buffer []uint8) string {
	str := string(buffer)
	base64Str := base64.StdEncoding.EncodeToString([]byte(str))

	// Convert standard Base64 to Base64 URL encoding
	base64Url := strings.NewReplacer("+", "-", "/", "_", "=", "").Replace(base64Str)

	return base64Url
}

func BufferToHex(buffer []uint8) string {
	hex := ""
	for _, byte := range buffer {
		hex += fmt.Sprintf("%02x", byte)
	}
	return hex
}

func EncryptMetadata(fileHeader *multipart.FileHeader, personalSignature string) (encryptedFilename string, encryptedFiletype string, err error) {
	// get aesKey, salt and iv

	//get salt
	//if salt is not provided, generate a random salt
	/*
		var salt []uint8
		var iv []uint8
		if salt == nil {
			salt, err = GenerateRandomBytes(aes.BlockSize)
			if err != nil {
				return
			}
		} else {
			salt = []uint8(salt)
		}
		if iv == nil {
			iv, err = GenerateRandomBytes(12)
			if err != nil {
				return
			}
		} else {
			iv = []uint8(iv)
		}
	*/

	// get aesKey, salt and iv

	//get salt
	//if salt is not provided, generate a random salt
	var salt []uint8
	var iv []uint8
	if salt == nil {
		//generate empty salt of length 16 and filled with 0
		salt = make([]uint8, 16)
		for i := range salt {
			salt[i] = 0
		}
	} else {
		salt = []uint8(salt)
	}
	if iv == nil {
		//generate empty iv of length 12 and filled with 0
		iv = make([]uint8, 12)
		for i := range iv {
			iv[i] = 0
		}
	} else {
		iv = []uint8(iv)
	}

	aesKey, salt, iv, err := GetAESKeySaltIV(personalSignature, salt, iv)
	if err != nil {
		return "", "", err
	}

	
	//encrypt file metadata
	fileNameBytes := []uint8(fileHeader.Filename)
	fileTypeBytes := []uint8(fileHeader.Header.Get("Content-Type"))
	
	encryptedFilenameBytes, err := EncryptBytes(fileNameBytes, aesKey, salt, iv)
	if err != nil {
		return "", "", err
	}
	encryptedFiletypeBytes, err := EncryptBytes(fileTypeBytes, aesKey, salt, iv)
	if err != nil {
		return "", "", err
	}

	//transform to base64 url
	encryptedFilename = BufferToBase64Url(encryptedFilenameBytes)
	encryptedFiletype = BufferToHex(encryptedFiletypeBytes)


	return encryptedFilename, encryptedFiletype, nil
}

func GenerateFileCID(fileBuffer multipart.File) (cid string, err error) {

	hasher := sha256.New()
	buf := make([]byte, 1<<20) // 1 MB buffer

	for {
		n, err := fileBuffer.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		hasher.Write(buf[:n])
	}

	hashSum := hasher.Sum(nil)
	// Create a multihash
	mhash, err := multihash.Encode(hashSum, multihash.SHA2_256)
	if err != nil {
		return "", err
	}

	// Create a CID
	c := cidUtil.NewCidV1(cidUtil.Raw, mhash)
	if err != nil {
		return "", err
	}

	return c.String(), nil
}

func GenerateCID(buffer []byte) {
	// Create a cid manually by specifying the 'prefix' parameters
	pref := cidUtil.Prefix{
		Version:  1,
		Codec:    cidUtil.Raw,
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	c, err := pref.Sum(buffer)
	if err != nil {
		panic(err)
	}

	log.Println("Created CID: ", c)
}

func EncryptFileBuffer(file *multipart.FileHeader) (cidOriginalStr, cidOfEncryptedBufferStr, encryptionTime string, encryptedFileBuffer []uint8, err error) {

	// get aesKey, salt and iv

	//get salt
	//if salt is not provided, generate a random salt
	var salt []uint8
	var iv []uint8
	if salt == nil {
		//generate empty salt of length 16 and filled with 0
		salt = make([]uint8, 16)
		for i := range salt {
			salt[i] = 0
		}
	} else {
		salt = []uint8(salt)
	}
	if iv == nil {
		//generate empty iv of length 12 and filled with 0
		iv = make([]uint8, 12)
		for i := range iv {
			iv[i] = 0
		}
	} else {
		iv = []uint8(iv)
	}

	if err != nil {
		return "", "", "", nil, err
	}

	// Read the contents of the file buffer
	fileBuffer, err := file.Open()
	if err != nil {
		err := fmt.Errorf("cannot open file: %s", err)
		return "", "", "", nil, err
	}
	defer fileBuffer.Close()
	//transform fileBuffer (multipart.File)) to []uint8

	cidOriginalStr, err = GenerateFileCID(fileBuffer)
	if err != nil {
		err = fmt.Errorf("cannot generate file CID: %s", err)
		return "", "", "", nil, err
	}

	// Reset the file pointer to the beginning of the file before reading it again
	if _, err := fileBuffer.Seek(0, 0); err != nil {
		return "", "", "", nil, fmt.Errorf("cannot seek file buffer: %s", err)
	}


	aesKey, salt, iv, err := GetAESKeySaltIV(cidOriginalStr, salt, iv)

	// Encrypt the file buffer
	start := time.Now()


	// Read the file buffer into a byte slice
	fileBytes, err := io.ReadAll(fileBuffer)
	if err != nil {
		err := fmt.Errorf("cannot read file buffer: %s", err)
		return "", "", "", nil, err
	}

	encryptedFileBufferBytes, err := EncryptBytes(fileBytes, aesKey, salt, iv)
	if err != nil {
		return "", "", "", nil, err
	}
	end := time.Now()

	encryptionTime = strconv.FormatInt(end.Sub(start).Milliseconds(), 36)

	// Convert the byte slice to ByteFile type
	byteFile := NewByteFile(encryptedFileBufferBytes)

	// get cidOfEncryptedBufferStr
	GenerateCID(encryptedFileBufferBytes)
	//log byteFile size
	cidOfEncryptedBufferStr, err = GenerateFileCID(byteFile)
	if err != nil {
		err = fmt.Errorf("cannot generate file CID: %s", err)
		return "", "", "", nil, err
	}
	byteFile.Close()

	return cidOriginalStr, cidOfEncryptedBufferStr, encryptionTime, encryptedFileBufferBytes, nil

}

type ByteFile struct {
	*bytes.Reader
}

func NewByteFile(data []byte) *ByteFile {
	return &ByteFile{
		Reader: bytes.NewReader(data),
	}
}

func (f *ByteFile) Close() error {
	// No-op for a byte slice, but necessary to implement multipart.File
	return nil
}

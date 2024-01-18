package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/ipfs/go-cid"
	mc "github.com/multiformats/go-multicodec"
	mh "github.com/multiformats/go-multihash"
)

// PublishFile creates a new public file.
func PublishFile(share_state entity.FileShareState, selectedShareFile form.CustomFileMeta) (*entity.PublicFile, error) {
	var publicFile entity.PublicFile

	publicFile.FileUID = share_state.FileUID
	publicFile.Name = selectedShareFile.Name
	publicFile.Mime = selectedShareFile.MimeType
	publicFile.Size = selectedShareFile.Size
	publicFile.CID = selectedShareFile.CID
	publicFile.CIDOriginalDecrypted = selectedShareFile.CIDOriginalEncrypted

	//ShareHash is the CID derived out of concatenated name, mime and size

	// Create a cid manually by specifying the 'prefix' parameters
	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(mc.Raw),
		MhType:   uint64(mh.SHA2_256),
		MhLength: -1,
	}

	// And then feed it with the data
	cid, err := pref.Sum([]byte(publicFile.Name + publicFile.Mime + string(rune(publicFile.Size))))
	if err != nil {
		return nil, err
	}

	publicFile.ShareHash = cid.String()

	err = db.UnscopedDb().Where("share_hash = ?", publicFile.ShareHash).First(&publicFile).Error
	if err == nil {
		db.UnscopedDb().Delete(&publicFile)
	}

	if err := publicFile.Create(); err != nil {
		return nil, err
	}

	return &publicFile, nil
}

func FindPublicFileByHash(shareHash string) (*entity.PublicFile, error) {
	var publicFile entity.PublicFile
	err := db.UnscopedDb().Where("share_hash = ?", shareHash).First(&publicFile).Error
	if err != nil {
		return nil, err
	}

	return &publicFile, nil
}

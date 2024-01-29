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


// PublishFileUserShared creates a new public file.
func PublishFileUserShared(share_state entity.FileShareStatesUserShared, selectedShareFile form.CustomFileMeta) (*entity.PublicFileUserShared, error) {
	var publicFile entity.PublicFileUserShared

	publicFile.FileUID = share_state.FileUID
	publicFile.Name = selectedShareFile.Name
	publicFile.Mime = selectedShareFile.MimeType
	publicFile.Size = selectedShareFile.Size
	publicFile.CID = selectedShareFile.CID
	publicFile.CIDOriginalDecrypted = selectedShareFile.CIDOriginalEncrypted

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

// GetFileShareStateByFileUIDAndUserID retrieves the FileShareStatesUserShared object and its associated PublicFile
// based on the provided fileUID and userID.
func GetFileShareStateByFileUIDAndUserID(fileUID string, userID uint) (*entity.FileShareStatesUserShared, error) {
    var fileShareState entity.FileShareStatesUserShared
    // Search for the sharing state by fileUID and userID
    result := db.Db().Preload("PublicFile").Where("file_uid = ? AND user_id = ?", fileUID, userID).First(&fileShareState)
    if result.Error != nil {
        return nil, result.Error
    }

    return &fileShareState, nil
}

// ConvertToDomainEntities converts the FileShareStatesUserShared and PublicFileUserShared objects
// to the desired FileShareState and PublicFile entities.
func ConvertToDomainEntities(fileShareStatesUserShared *entity.FileShareStatesUserShared) (entity.FileShareState) {
    fileShareState := entity.FileShareState{
		ID: 	1, // if the ID is not set, it will be 0, the entire sharestate wont be able to be used in frontend
        FileUID:    fileShareStatesUserShared.FileUID,
        PublicFile: entity.PublicFile{
            FileUID:              fileShareStatesUserShared.PublicFile.FileUID,
            ShareHash:            fileShareStatesUserShared.PublicFile.ShareHash,
            Name:                 fileShareStatesUserShared.PublicFile.Name,
            Mime:                 fileShareStatesUserShared.PublicFile.Mime,
            Size:                 fileShareStatesUserShared.PublicFile.Size,
            CID:                  fileShareStatesUserShared.PublicFile.CID,
            CIDOriginalDecrypted: fileShareStatesUserShared.PublicFile.CIDOriginalDecrypted,
        },
    }
    return fileShareState
}

// DeleteFileShareStatesUserShared deletes the sharing state of a file shared with a specific user based on its UID.
func DeleteFileShareStatesUserShared(fileUID string, userID uint) {
	var fileShareState entity.FileShareStatesUserShared
	// Search for the sharing state by the file UID and user ID
	result := db.Db().Unscoped().Where("file_uid = ? AND user_id = ?", fileUID, userID).First(&fileShareState)
	if result.Error == nil {
		db.Db().Delete(&fileShareState.PublicFile)
		db.Db().Unscoped().Delete(&fileShareState)
	}
}

func CreateShareStateUserShared(file *entity.File, userID uint) (file_share_state entity.FileShareStatesUserShared, err error) {
	var file_share_states = entity.FileShareStatesUserShared{
		FileUID: file.UID,
		UserID:  userID,
	}

	if err := db.Db().Create(&file_share_state).Error; err != nil {
		return file_share_states, err
	}

	return file_share_states, nil
}
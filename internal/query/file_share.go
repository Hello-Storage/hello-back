package query

import (
	"github.com/Hello-Storage/hello-back/internal/db"
	"github.com/Hello-Storage/hello-back/internal/entity"
	"github.com/Hello-Storage/hello-back/internal/form"
	"github.com/ipfs/go-cid"
	mc "github.com/multiformats/go-multicodec"
	mh "github.com/multiformats/go-multihash"
	"gorm.io/gorm"
)

// PublishFile creates a new public file.
func PublishFile(tx *gorm.DB, share_state entity.FileShareState, selectedShareFile form.CustomFileMeta) (*entity.PublicFile, error) {
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

	err = tx.Unscoped().Where("share_hash = ?", publicFile.ShareHash).First(&publicFile).Error
	if err == nil {
		tx.Unscoped().Delete(&publicFile)
	}

	if err := publicFile.TxCreate(tx); err != nil {
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
func PublishFileUserShared(tx *gorm.DB, share_state entity.FileShareStatesUserShared, selectedShareFile form.CustomFileMeta) (*entity.PublicFileUserShared, error) {
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

	err = tx.Unscoped().Where("share_hash = ?", publicFile.ShareHash).First(&publicFile).Error
	if err == nil {
		if err = tx.Unscoped().Delete(&publicFile).Error; err != nil {
			return nil, err
		}
	}

	if err := publicFile.TxCreate(tx); err != nil {
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
func ConvertToDomainEntities(fileShareStatesUserShared *entity.FileShareStatesUserShared) entity.FileShareState {
	fileShareState := entity.FileShareState{
		ID:      1, // if the ID is not set, it will be 0, the entire sharestate wont be able to be used in frontend
		FileUID: fileShareStatesUserShared.FileUID,
		PublicFile: entity.PublicFile{
			FileUID:              fileShareStatesUserShared.PublicFileUserShared.FileUID,
			ShareHash:            fileShareStatesUserShared.PublicFileUserShared.ShareHash,
			Name:                 fileShareStatesUserShared.PublicFileUserShared.Name,
			Mime:                 fileShareStatesUserShared.PublicFileUserShared.Mime,
			Size:                 fileShareStatesUserShared.PublicFileUserShared.Size,
			CID:                  fileShareStatesUserShared.PublicFileUserShared.CID,
			CIDOriginalDecrypted: fileShareStatesUserShared.PublicFileUserShared.CIDOriginalDecrypted,
		},
	}
	return fileShareState
}

// DeleteFileShareStatesUserShared deletes the sharing state of a file based on its UID.
func DeleteFileShareStatesUserShared(tx *gorm.DB, fileUID string, userID uint) string {
	var fileShareState entity.FileShareStatesUserShared
	var filepublicf entity.PublicFileUserShared
	var cidOriginalDecrypted string

	result := tx.Unscoped().Where("file_uid = ? AND user_id = ?", fileUID, userID).First(&fileShareState)
	if result.Error == nil {
		if err := tx.Unscoped().Delete(&fileShareState.PublicFileUserShared).Error; err != nil {
			log.Errorf("Error deleting public file user shared: %v", err)
		}
		if err := tx.Unscoped().Delete(&fileShareState).Error; err != nil {
			log.Errorf("Error deleting file share state user shared: %v", err)
		}
		cidOriginalDecrypted = fileShareState.PublicFileUserShared.CIDOriginalDecrypted
	}
	if err := tx.Unscoped().Where("file_uid = ?", fileUID).Delete(&filepublicf).Error; err != nil {
		log.Errorf("Error deleting orphan PublicFileUserShared records: %v", err)
	}

	return cidOriginalDecrypted
}

// DeleteFileShareState deletes the sharing state of a file based on its UID.
func DeleteFileShareState(tx *gorm.DB, fileUID string) {
	var fileShareState entity.FileShareState
	// Search for the sharing state by the file UID
	result := tx.Unscoped().Where("file_uid = ?", fileUID).First(&fileShareState)
	if result.Error == nil {
		tx.Unscoped().Delete(&fileShareState.PublicFile)
		tx.Unscoped().Delete(&fileShareState)
	}
}

func CreateShareStateUserShared(tx *gorm.DB, file *entity.File, userID uint) (fileShareState entity.FileShareStatesUserShared, err error) {
	fileShareState = entity.FileShareStatesUserShared{
		FileUID: file.UID,
		UserID:  userID,
	}

	if err := fileShareState.TxCreate(tx); err != nil {
		return fileShareState, err
	}

	return fileShareState, nil
}

package form

import (
	"github.com/Hello-Storage/hello-back/internal/entity"
)

type CreateFolder struct {
	Title string `json:"title"`
	Root  string `json:"root"`
	EncryptionStatus    entity.EncryptionStatus `json:"encryption_status"`
}

type UpdateFolder struct {
	Id   string `json:"id"`
	Uid  string `json:"uid"`
	Root string `json:"root"`
}

type FolderContent struct {
	Folder SharedFolder `json:"folder"`
}

type SharedFolder struct {
	Uid  string `json:"uid"`
	Title string `json:"title"`
	Files []CustomFileMeta `json:"files"`
	Folders []FolderContent `json:"folders"`
}

type SharedFolderRes struct {
	Folder entity.Folder `json:"folder"`
}

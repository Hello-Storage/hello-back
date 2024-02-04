package form

import "github.com/Hello-Storage/hello-back/internal/entity"

type BaseMeta struct {
	ID               int                     `json:"id"`
	CreatedAt        string                  `json:"created_at"`
	DeletedAt        *string                 `json:"deleted_at,omitempty"`
	UpdatedAt        string                  `json:"updated_at"`
	EncryptionStatus entity.EncryptionStatus `json:"encryption_status"`
	Decrypted        *bool                   `json:"decrypted,omitempty"`
}

type CustomFileMeta struct {
	BaseMeta
	UID                     string  `json:"uid"`
	CID                     string  `json:"cid"`
	CIDOriginalEncrypted    string  `json:"cid_original_encrypted"`
	CIDOriginalEncryptedB64 *string `json:"cid_original_encrypted_base64_url,omitempty"`
	Name                    string  `json:"name"`
	Root                    string  `json:"root"`
	MimeType                string  `json:"mime_type"`
	Size                    int64   `json:"size"`
	MediaType               string  `json:"media_type"`
	Path                    string  `json:"path"`
	Data                    *string `json:"data,omitempty"`
	IsOwner                 bool    `json:"isOwner"`
}

type FileResponse struct {
	ID                   uint                    `json:"id"`
	Name                 string                  `json:"name"`
	UID                  string                  `json:"uid"`
	Root                 string                  `json:"root"`
	CID                  string                  `json:"cid"`
	CIDOriginalEncrypted *string                 `json:"cid_original_encrypted"`
	Mime                 string                  `json:"mime"`
	Size                 int64                   `json:"size"`
	EnryptionStatus      entity.EncryptionStatus `json:"encryption_status"`
	IsInPool             *bool                   `json:"is_in_pool"`
	CreatedAt            string                  `json:"created_at"`
	UpdatedAt            string                  `json:"updated_at"`
}

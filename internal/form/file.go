package form

type UpdateFileRoot struct {
	Uid  string `json:"uid"`
	Root string `json:"root"`
	Id   string `json:"id"`
}

type UpdateFileIpfsHash struct {
	IpfsHash string `json:"ipfs_hash"`
	Uid      string `json:"uid"`
}
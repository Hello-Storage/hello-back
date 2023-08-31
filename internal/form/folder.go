package form

type CreateFolder struct {
	Title string `json:"title"`
	Root  string `json:"root"`
}

type UpdateFolder struct {
	Id   string `json:"id"`
	Uid  string `json:"uid"`
	Root string `json:"root"`
}

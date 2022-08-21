package video

type H5sResponse struct {
	BStatus     bool   `json:"bStatus"`
	StrCode     string `json:"strCode"`
	StrFileName string `json:"strFileName"`
	StrUrl      string `json:"strUrl"`
	Record      []struct {
		StrPath string `json:"strPath"`
	} `json:"record"`
}

package response

type S3Response struct {
	Filename string `json:"filename"`
	Success  bool   `json:"success"`
	Filepath string `json:"filepath"`
	Key      string `json:"key"`
	Message  string `json:"message"`
	Order    int    `json:"order"`
}

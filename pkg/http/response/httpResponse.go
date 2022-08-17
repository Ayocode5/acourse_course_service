package response

type HttpResponse struct {
	StatusCode int         `json:"status_code,omitempty"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

type HttpPaginationResponse struct {
	PerPage int64 `json:"per_page"`
	Page    int64 `json:"page"`
	HttpResponse
}

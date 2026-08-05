package v1

// DataResp represents a standard response structure wrapping payload data.
type DataResp struct {
	Data any `json:"data"`
}

// FullURLResp represents the response payload containing the original full URL.
type FullURLResp struct {
	URL string `json:"url"`
}

// ShortURLResp represents the response payload containing the generated short URL.
type ShortURLResp struct {
	ShortURL string `json:"short_url"`
}

// CreateShortURLReq represents the incoming JSON request payload to create a short URL.
type CreateShortURLReq struct {
	URL string `json:"url" binding:"required"`
}

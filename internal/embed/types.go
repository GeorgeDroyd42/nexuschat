package embed

type EmbedData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	Success     bool   `json:"success"`
}

type EmbedResponse struct {
	Success bool        `json:"success"`
	Data    *EmbedData  `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
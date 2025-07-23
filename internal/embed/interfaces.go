package embed

import (
	"net/http"
)

type EmbedService interface {
	GetEmbed(url string) (*EmbedData, error)
	ProxyImage(url string) (*http.Response, error)
	ValidateURL(url string) bool
	ExtractMetaTags(html string) *EmbedData
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type CacheProvider interface {
	Get(key string, dest interface{}) (bool, error)
	Set(key string, value interface{}, duration interface{}) error
}
package embed

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	"auth.com/v4/cache"
	"auth.com/v4/utils"
	"github.com/labstack/echo/v4"
)

var (
	ogTitleRegex       = regexp.MustCompile(`<meta\s+property="og:title"\s+content="([^"]*)"`)
	ogDescriptionRegex = regexp.MustCompile(`<meta\s+property="og:description"\s+content="([^"]*)"`)
	ogImageRegex       = regexp.MustCompile(`<meta\s+property="og:image"\s+content="([^"]*)"`)
	ogSiteNameRegex    = regexp.MustCompile(`<meta\s+property="og:site_name"\s+content="([^"]*)"`)
	titleRegex         = regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)
)

func ValidateURL(url string) bool {
	if url == "" {
		return false
	}
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func GetEmbedHandler(c echo.Context) error {
	if _, err := utils.RequireUserID(c); err != nil {
		return err
	}

	url := c.QueryParam("url")
	if !ValidateURL(url) {
		return utils.SendErrorResponse(c, utils.ErrUnauthorized)
	}

	cacheKey := CacheKeyPrefix + url
	var cachedResult map[string]interface{}
	if found, err := cache.Provider.Get(cacheKey, &cachedResult); err == nil && found {
		return c.JSON(200, cachedResult)
	}

	client := createHTTPClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return c.JSON(200, map[string]interface{}{"success": false})
	}

	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(200, map[string]interface{}{"success": false})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.JSON(200, map[string]interface{}{"success": false})
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(200, map[string]interface{}{"success": false})
	}

	embed := extractMetaTags(string(body))
	embed["success"] = true

	cache.Provider.Set(cacheKey, embed, CacheDuration)

	return c.JSON(200, embed)
}

func extractMetaTags(html string) map[string]interface{} {
	result := make(map[string]interface{})

	if matches := ogTitleRegex.FindStringSubmatch(html); len(matches) > 1 {
		result["title"] = matches[1]
	}
	if matches := ogDescriptionRegex.FindStringSubmatch(html); len(matches) > 1 {
		result["description"] = matches[1]
	}
	if matches := ogImageRegex.FindStringSubmatch(html); len(matches) > 1 {
		result["image"] = "/api/proxy-image?url=" + matches[1]
	}
	if matches := ogSiteNameRegex.FindStringSubmatch(html); len(matches) > 1 {
		result["site_name"] = matches[1]
	}

	if result["title"] == nil {
		if matches := titleRegex.FindStringSubmatch(html); len(matches) > 1 {
			result["title"] = strings.TrimSpace(matches[1])
		}
	}

	return result
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: RequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= MaxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func ProxyImageHandler(c echo.Context) error {
	if _, err := utils.RequireUserID(c); err != nil {
		return err
	}

	imageURL := c.QueryParam("url")
	if !ValidateURL(imageURL) {
		return c.NoContent(404)
	}

	client := createHTTPClient()
	req, _ := http.NewRequest("GET", imageURL, nil)
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return c.NoContent(404)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.NoContent(404)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return c.NoContent(404)
	}

	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

	_, err = io.Copy(c.Response().Writer, resp.Body)
	return err
}
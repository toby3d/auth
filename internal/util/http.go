package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/domain"
)

var ErrEndpointNotExist = domain.NewError(
	domain.ErrorCodeServerError,
	"cannot found any endpoints",
	"https://indieauth.net/source/#discovery-0",
)

func ExtractEndpoints(resp *http.Response, rel string) []*domain.URL {
	results := make([]*domain.URL, 0)

	urls, err := ExtractEndpointsFromHeader(resp, rel)
	if err == nil {
		results = append(results, urls...)
	}

	urls, err = ExtractEndpointsFromBody(resp, rel)
	if err == nil {
		results = append(results, urls...)
	}

	return results
}

func ExtractEndpointsFromHeader(resp *http.Response, rel string) ([]*domain.URL, error) {
	results := make([]*domain.URL, 0)

	for _, link := range linkheader.Parse(string(resp.Header.Peek(http.HeaderLink))) {
		if !strings.EqualFold(link.Rel, rel) {
			continue
		}

		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(link.URL)); err != nil {
			return nil, fmt.Errorf("cannot parse header endpoint: %w", err)
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractEndpointsFromBody(resp *http.Response, rel string) ([]*domain.URL, error) {
	endpoints, ok := microformats.Parse(bytes.NewReader(resp.Body()), nil).Rels[rel]
	if !ok || len(endpoints) == 0 {
		return nil, ErrEndpointNotExist
	}

	results := make([]*domain.URL, 0)

	for i := range endpoints {
		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(endpoints[i])); err != nil {
			return nil, fmt.Errorf("cannot parse body endpoint: %w", err)
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractMetadata(resp *http.Response, client *http.Client) (*domain.Metadata, error) {
	endpoints := ExtractEndpoints(resp, "indieauth-metadata")
	if len(endpoints) == 0 {
		return nil, ErrEndpointNotExist
	}

	_, body, err := client.Get(nil, endpoints[len(endpoints)-1].String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata endpoint configuration: %w", err)
	}

	result := new(domain.Metadata)
	if err = json.Unmarshal(body, result); err != nil {
		return nil, fmt.Errorf("cannot unmarshal emtadata configuration: %w", err)
	}

	return result, nil
}

func ExtractProperty(resp *http.Response, itemType, key string) []interface{} {
	//nolint: exhaustivestruct // only Host part in url.URL is needed
	data := microformats.Parse(bytes.NewReader(resp.Body()), &url.URL{
		Host: string(resp.Header.Peek(http.HeaderHost)),
	})

	return findProperty(data.Items, itemType, key)
}

func contains(src []string, find string) bool {
	for i := range src {
		if !strings.EqualFold(src[i], find) {
			continue
		}

		return true
	}

	return false
}

func findProperty(src []*microformats.Microformat, itemType, key string) []interface{} {
	for _, item := range src {
		if contains(item.Type, itemType) {
			return item.Properties[key]
		}

		result := findProperty(item.Children, itemType, key)
		if result == nil {
			continue
		}

		return result
	}

	return nil
}

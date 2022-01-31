package util

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	http "github.com/valyala/fasthttp"
	"willnorris.com/go/microformats"

	"source.toby3d.me/website/indieauth/internal/domain"
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
			return nil, err
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractEndpointsFromBody(resp *http.Response, rel string) ([]*domain.URL, error) {
	endpoints, ok := microformats.Parse(bytes.NewReader(resp.Body()), nil).Rels[rel]
	if !ok || len(endpoints) == 0 {
		return nil, nil
	}

	results := make([]*domain.URL, 0)

	for i := range endpoints {
		u := http.AcquireURI()
		if err := u.Parse(resp.Header.Peek(http.HeaderHost), []byte(endpoints[i])); err != nil {
			return nil, err
		}

		results = append(results, &domain.URL{URI: u})
	}

	return results, nil
}

func ExtractMetadata(resp *http.Response, client *http.Client) (*domain.Metadata, error) {
	endpoints := ExtractEndpoints(resp, "indieauth-metadata")
	if len(endpoints) == 0 {
		return nil, nil
	}

	_, body, err := client.Get(nil, endpoints[len(endpoints)-1].String())
	if err != nil {
		return nil, err
	}

	result := new(domain.Metadata)
	if err = json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}

func ExtractProperty(resp *http.Response, t, key string) []interface{} {
	data := microformats.Parse(bytes.NewReader(resp.Body()), &url.URL{
		Host: string(resp.Header.Peek(http.HeaderHost)),
	})

	for _, item := range data.Items {
		if !contains(item.Type, t) {
			continue
		}

		return item.Properties[key]
	}

	return nil
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

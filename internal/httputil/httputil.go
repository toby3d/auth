package httputil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/goccy/go-json"
	"github.com/tomnomnom/linkheader"
	"golang.org/x/exp/slices"
	"willnorris.com/go/microformats"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/domain"
)

const RelIndieauthMetadata = "indieauth-metadata"

var ErrEndpointNotExist = domain.NewError(
	domain.ErrorCodeServerError,
	"cannot found any endpoints",
	"https://indieauth.net/source/#discovery-0",
)

func ExtractFromMetadata(client *http.Client, u string) (*domain.Metadata, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)

	endpoints := ExtractEndpoints(buf, resp.Request.URL, resp.Header.Get(common.HeaderLink), RelIndieauthMetadata)
	if len(endpoints) == 0 {
		return nil, ErrEndpointNotExist
	}

	if resp, err = client.Get(endpoints[len(endpoints)-1].String()); err != nil {
		return nil, fmt.Errorf("failed to fetch metadata endpoint configuration: %w", err)
	}

	result := new(domain.Metadata)
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("cannot unmarshal emtadata configuration: %w", err)
	}

	return result, nil
}

func ExtractEndpoints(body io.Reader, u *url.URL, linkHeader, rel string) []*url.URL {
	results := make([]*url.URL, 0)

	urls, err := ExtractEndpointsFromHeader(linkHeader, rel)
	if err == nil {
		results = append(results, urls...)
	}

	urls, err = ExtractEndpointsFromBody(body, u, rel)
	if err == nil {
		results = append(results, urls...)
	}

	return results
}

func ExtractEndpointsFromHeader(linkHeader, rel string) ([]*url.URL, error) {
	results := make([]*url.URL, 0)

	for _, link := range linkheader.Parse(linkHeader) {
		if !strings.EqualFold(link.Rel, rel) {
			continue
		}

		u, err := url.Parse(link.URL)
		if err != nil {
			return nil, fmt.Errorf("cannot parse header endpoint: %w", err)
		}

		results = append(results, u)
	}

	return results, nil
}

func ExtractEndpointsFromBody(body io.Reader, u *url.URL, rel string) ([]*url.URL, error) {
	endpoints, ok := microformats.Parse(body, u).Rels[rel]
	if !ok || len(endpoints) == 0 {
		return nil, ErrEndpointNotExist
	}

	results := make([]*url.URL, 0)

	for i := range endpoints {
		u, err := url.Parse(endpoints[i])
		if err != nil {
			return nil, fmt.Errorf("cannot parse body endpoint: %w", err)
		}

		results = append(results, u)
	}

	return results, nil
}

func ExtractProperty(body io.Reader, u *url.URL, itemType, key string) []any {
	if data := microformats.Parse(body, u); data != nil {
		return FindProperty(data.Items, itemType, key)
	}

	return nil
}

func FindProperty(src []*microformats.Microformat, itemType, key string) []any {
	for _, item := range src {
		if slices.Contains(item.Type, itemType) {
			return item.Properties[key]
		}

		if result := FindProperty(item.Children, itemType, key); result != nil {
			return result
		}
	}

	return nil
}

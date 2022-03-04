// Package keybase provides a client for the Keybase API.
//
// The Keybase API is documented at https://keybase.io/docs/api/1.0/.
package keybase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiScheme      = "https"
	apiHost        = "keybase.io"
	userLookupPath = "/_/api/1.0/user/lookup.json"
)

// GetPictureByKeySuffix returns the primary picture of a user queried by suffix
// of their public key.
func GetPictureByKeySuffix(c context.Context, keySuffix string) (string, error) {
	data, err := UserLookup(c, UserLookupQuery{KeySuffix: keySuffix}, []string{"pictures"})
	if err != nil {
		return "", err
	}

	return data.Them[0].Pictures.Primary.URL, nil
}

type UserLookupQuery struct {
	KeySuffix string
}

// UserLookup performs a user lookup as described by the Keybase API here:
// https://keybase.io/docs/api/1.0/call/user/lookup.
func UserLookup(c context.Context, query UserLookupQuery, fields []string) (UserLookupResponse, error) {
	var data UserLookupResponse

	q := make(url.Values)
	q.Add("fields", strings.Join(fields, ","))
	if len(query.KeySuffix) > 0 {
		q.Add("key_suffix", query.KeySuffix)
	}

	u := url.URL{
		Scheme:   apiScheme,
		Host:     apiHost,
		Path:     userLookupPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequestWithContext(c, "GET", u.String(), nil)
	if err != nil {
		return data, fmt.Errorf("preparing request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return data, fmt.Errorf("performing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return data, fmt.Errorf("http status code %d", res.StatusCode)
	}

	dec := json.NewDecoder(res.Body)

	err = dec.Decode(&data)
	if err != nil {
		return data, fmt.Errorf("decoding response: %w", err)
	}

	if data.Status.Code != 0 {
		return data, fmt.Errorf("api error: status=%v name=%v desc=%v", data.Status.Code, data.Status.Name, data.Status.Desc)
	}

	if len(data.Them) == 0 {
		return data, fmt.Errorf("no user matched query")
	}

	if len(data.Them) > 1 {
		return data, fmt.Errorf("more than one user matched the query")
	}

	return data, nil
}

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type UserLookupResponse struct {
	Status Status `json:"status,omitempty"`
	Them   []User `json:"them,omitempty"`
}

type Status struct {
	Code int    `json:"code"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type User struct {
	ID       string   `json:"id"`
	Pictures Pictures `json:"pictures"`
}

type Pictures struct {
	Primary Picture `json:"primary"`
}

type Picture struct {
	URL    string  `json:"url"`
	Source *string `json:"source"`
}

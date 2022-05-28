package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Lambels/relationer/internal"
	"github.com/Lambels/relationer/internal/rest"
)

// Client is an http client which implements the internal.GraphStore
type Client struct {
	*http.Client

	URL string
}

func NewClient(client *http.Client, base string) *Client {
	return &Client{client, base}
}

func (c *Client) AddPerson(ctx context.Context, person *internal.Person) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&person); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL+"/people",
		&buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusCreated {
		return parseRespErr(resp)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemovePerson(ctx context.Context, id int64) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.URL+"/people/"+fmt.Sprint(id),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusNoContent {
		return parseRespErr(resp)
	}
	return resp.Body.Close()
}

func (c *Client) GetPerson(ctx context.Context, id int64) (*internal.Person, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.URL+"/people/"+fmt.Sprint(id),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusOK {
		return nil, parseRespErr(resp)
	}
	defer resp.Body.Close()

	var person internal.Person
	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, err
	}

	return &person, nil
}

func (c *Client) AddFriendship(ctx context.Context, friendship internal.Friendship) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(friendship); err != nil {
		return internal.WrapError(err, internal.EINTERNAL, "json.Encode")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.URL+"/friendship",
		&buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusCreated {
		return parseRespErr(resp)
	}
	return resp.Body.Close()
}

func (c *Client) GetDepth(ctx context.Context, id1, id2 int64) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.URL+"/friendship/depth/"+fmt.Sprint(id1)+"/"+fmt.Sprint(id2),
		nil,
	)
	if err != nil {
		return -1, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return -1, internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusOK {
		return -1, parseRespErr(resp)
	}
	defer resp.Body.Close()

	var depth rest.GetDepthResponse
	if err := json.NewDecoder(resp.Body).Decode(&depth); err != nil {
		return -1, err
	}

	return depth.Depth, nil
}

func (c *Client) GetFriendship(ctx context.Context, id int64) (internal.Friendship, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.URL+"/friendship/"+fmt.Sprint(id),
		nil,
	)
	if err != nil {
		return internal.Friendship{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return internal.Friendship{}, internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusOK {
		return internal.Friendship{}, parseRespErr(resp)
	}
	defer resp.Body.Close()

	var friendship internal.Friendship
	if err := json.NewDecoder(resp.Body).Decode(&friendship); err != nil {
		return friendship, err
	}

	return friendship, nil
}

func (c *Client) GetAll(ctx context.Context) ([]internal.Friendship, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.URL+"/people/",
		nil,
	)
	if err != nil {
		return []internal.Friendship{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return []internal.Friendship{}, internal.WrapError(err, internal.ECONFLICT, "c.Do")
	} else if resp.StatusCode != http.StatusOK {
		return []internal.Friendship{}, parseRespErr(resp)
	}
	defer resp.Body.Close()

	var friendship []internal.Friendship
	if err := json.NewDecoder(resp.Body).Decode(&friendship); err != nil {
		return friendship, err
	}

	return friendship, nil
}

// parseRespErr parses a json error from the response to a *internal.Error.
//
// will close the resp.Body
func parseRespErr(resp *http.Response) error {
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var internalErr rest.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&internalErr); err != nil {
		msg := string(buf) // try to read message.
		if msg == "" {
			msg = "server responded with no message"
		}
		return internal.Errorf(internal.ECodeFromStatusCode(resp.StatusCode), msg)
	}
	return internal.Errorf(internal.ECodeFromStatusCode(resp.StatusCode), internalErr.Error)
}

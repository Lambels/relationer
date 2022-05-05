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

type Client struct {
	http.Client

	URL string
}

func NewClient(client http.Client, base string) *Client {
	return &Client{client, base}
}

func (c *Client) AddPerson(ctx context.Context, person *internal.Person) error {
	var buf *bytes.Buffer
	if err := json.NewEncoder(buf).Encode(person); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/person", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
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
	var buf *bytes.Buffer
	if err := json.NewEncoder(buf).Encode(rest.RemovePersonRequest{Id: id}); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		c.URL+"/person/"+fmt.Sprint(id),
		buf,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusCreated {
		return parseRespErr(resp)
	}
	resp.Body.Close()
	return nil
}

func (c *Client) GetPerson(ctx context.Context, id int64) (*internal.Person, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.URL+"/person/"+fmt.Sprint(id),
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusCreated {
		return nil, parseRespErr(resp)
	}
	defer resp.Body.Close()

	var person internal.Person
	if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
		return nil, err
	}

	return &person, nil
}

func (c *Client) AddFriendship(context.Context, internal.Friendship) error {

}

func (c *Client) GetDepth(context.Context, int64, int64) (int, error) {

}

func (c *Client) GetFriendship(context.Context, int64) (internal.Friendship, error) {

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

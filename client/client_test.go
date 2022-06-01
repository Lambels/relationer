package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lambels/relationer/internal"
)

func TestAddPerson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(buf, &data); err != nil {
			t.Fatal(err)
		}

		if got, want := data["name"], "foo"; got != want {
			t.Fatalf("Got: %v Want: %v", got, want)
		}

		data["id"] = 1

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)
	}))
	defer ts.Close()

	c := New(&ClientConfig{
		URL:            ts.URL,
		Client:         http.DefaultClient,
		ConsumerConfig: nil,
	})
	id, err := c.AddPerson(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := id, int64(1); got != want {
		t.Fatalf("Got: %v Want: %v", got, want)
	}
}

func TestAddFriendship(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		var data internal.Friendship
		if err := json.Unmarshal(buf, &data); err != nil {
			t.Fatal(err)
		}

		if got, want := data.P1.ID, int64(1); got != want {
			t.Fatalf("Got: %v Want: %v", got, want)
		}

		if got, want := data.With[0], int64(2); got != want {
			t.Fatalf("Got: %v Want: %v", got, want)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := New(&ClientConfig{
		URL:            ts.URL,
		Client:         http.DefaultClient,
		ConsumerConfig: nil,
	})
	if err := c.AddFriendship(context.Background(), 1, 2); err != nil {
		t.Fatal(err)
	}
}

func TestCtxCancelDetached(t *testing.T) {
	c := New(nil)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	recv, err := c.StartListenDetached(ctx)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-waitRecvClose(recv):
	case <-time.After(2 * time.Second):
		t.Fatal("recv channel didnt close")
	}
}

func TestCtxCancelAttached(t *testing.T) {
	c := New(nil)
	defer c.Close()

	recv1, err := c.StartListenDetached(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	recv2, err := c.StartListenAttached(ctx, false)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-waitRecvClose(recv2):
		if recvClosed(recv1) {
			t.Fatal("root reciever closed")
		}
	case <-waitRecvClose(recv1):
		t.Fatal("root reciever closed")
	case <-time.After(2 * time.Second):
		t.Fatal("no reciever closed")
	}
}

func TestAttachedLast(t *testing.T) {
	c := New(nil)
	defer c.Close()

	_, err := c.StartListenDetached(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.StartListenDetached(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.StartListenAttached(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.StartListenAttached(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(c.consumerPool[1].notify), 3; got != want {
		t.Fatalf("Got: %v Want: %v", got, want)
	}
}

func TestAttachedLastNoCons(t *testing.T) {
	c := New(nil)
	defer c.Close()

	if _, err := c.StartListenAttached(context.Background(), true); err == nil {
		t.Fatal("expected error")
	}
}

func TestCtxCancelDetachedAll(t *testing.T) {
	c := New(nil)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	recv1, err := c.StartListenDetached(ctx)
	if err != nil {
		t.Fatal(err)
	}

	recv2, err := c.StartListenAttached(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-waitRecvClose(recv1):
		if !recvClosed(recv2) {
			t.Fatal("attached reciever not closed")
		}
	case <-time.After(2 * time.Second):
	}
}

func waitRecvClose(recv <-chan *Message) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		for range recv {
		}

		ch <- struct{}{}
	}()
	return ch
}

func recvClosed(recv <-chan *Message) bool {
	select {
	case _, ok := <-recv:
		return !ok
	case <-time.After(1e9):
		return false
	}
}

package consul

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

func HttpDo(ctx context.Context, f func(*http.Response, error) error) error {
	tr := http.Transport{}
	client := http.Client{
		Transport: tr,
	}

	c := make(chan error, 1)

	go func() {
		c <- f(client.Do(req))
	}()

	select {
	case <-ctx.Done():
		tr.CancelRequest(req)

		// wait f
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}

}

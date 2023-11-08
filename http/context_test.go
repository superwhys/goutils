package http

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestClientContext(t *testing.T) {
	t.Run("testContextFetch", func(t *testing.T) {
		resp, err := Cli.Start().
			SetMethod("GET").
			SetURL(fmt.Sprintf("%v/%v", srvApi, "test_get")).
			SetHeaders(DefaultJsonHeader()).
			Fetch(context.Background()).
			BodyString()
		assert.Nil(t, err)
		assert.Equal(t, `{"message":"do success"}`, resp)
	})
}

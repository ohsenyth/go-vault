package vault

import (
	"context"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
)

func NewHTTPServer(ctx context.Context, endpoints Endpoints) http.Handler {
	m := http.NewServeMux()
	m.Handle("/hash", httptransport.NewServer(
		ctx,
		endpoints.HashEndpoint,
		decodeHashRequest,
		endcodeResponse,
	))
	m.Handle("/validate", httptransport.NewServer(
		ctx,
		endpoints.ValidateEndpoint,
		decodeValidateRequest,
		endcodeResponse,
	))
	return m
}

package addsvc

// This file provides server-side bindings for the HTTP transport.
// It utilizes the transport/http.Server.

import (
	//"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	//"io/ioutil"
	"net/http"

	//"os"
	"strings"
	//stdopentracing "github.com/opentracing/opentracing-go"
	//"github.com/go-kit/kit/log"
	//"github.com/go-kit/kit/tracing/opentracing"
	//httptransport "github.com/go-kit/kit/transport/http"
)

// Business-domain errors like these may be served in two ways: returned
// directly by endpoints, or bundled into the response struct. Both methods can
// be made to work, but errors returned directly by endpoints are counted by
// middlewares that check errors, like circuit breakers.
//
// If you don't want that behavior -- and you probably don't -- then it's better
// to bundle errors into the response struct.

var (
	// ErrTwoZeroes is an arbitrary business rule for the Add method.
	ErrTwoZeroes = errors.New("can't sum two zeroes")

	// ErrIntOverflow protects the Add method. We've decided that this error
	// indicates a misbehaving service and should count against e.g. circuit
	// breakers. So, we return it directly in endpoints, to illustrate the
	// difference. In a real service, this probably wouldn't be the case.
	ErrIntOverflow = errors.New("integer overflow")

	// ErrMaxSizeExceeded protects the Concat method.
	ErrMaxSizeExceeded = errors.New("result exceeds maximum size")
)

func formatRequest(r *http.Request) string {
	// Create return string
	var reqstr []string

	// Add the method etc.
	reqstr = append(reqstr, fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto))

	// Add the host
	reqstr = append(reqstr, fmt.Sprintf("Host: %v", r.Host))

	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			reqstr = append(reqstr, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		reqstr = append(reqstr, "\n")
		reqstr = append(reqstr, r.Form.Encode())
	}

	// Return the reqstr as a string
	return strings.Join(reqstr, "\n")
}

//	"method", request.Method, "url", request.URL, "proto", request.Proto, "content-length", request.ContentLength,
//	"Host", request.Host, "RemoteAddr", request.RemoteAddr, "RequestURI", request.RequestURI,

// Utility Function - Log HTTP request
func getRequestInfoArgs(req *http.Request) []interface{} {
	return []interface{}{"tag", "#transport", "level", "debug", "method", req.Method, "url", req.URL, "proto", req.Proto}
}

func getRequestAdvancedInfoArgs(req *http.Request) []interface{} {
	return []interface{}{
		"level", "debug", "content-length", req.ContentLength, "Host", req.Host, "Header", req.Header,
		"RemoteAddr", req.RemoteAddr, "RequestURI", req.RequestURI,
	}
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	msg := err.Error()

	switch err {
	case ErrTwoZeroes, ErrMaxSizeExceeded, ErrIntOverflow:
		code = http.StatusBadRequest
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorWrapper{Error: msg})
}

func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

type errorWrapper struct {
	Error string `json:"error"`
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

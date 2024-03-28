// Package stdhttp provides utilities to simplify building HTTP and RESTful services.
package stdhttp

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
)

// RespondString responds on the given response with the status code and text body.  If an error occurs while responding
// it is journaled unless it is a client error.
func RespondString(ctx context.Context, writer http.ResponseWriter, status int, body string) {
	writer.WriteHeader(status)
	if _, err := writer.Write([]byte(body)); err != nil {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, "failed to write response")
		span.AddEvent("failed to write response")
		span.RecordError(err)
	}
}

// InternalError provides the client with a 500 and a "Internal Error" response.
func InternalError(writer http.ResponseWriter, request *http.Request, problem error) {
	ctx := request.Context()

	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, "internal server error")
	span.RecordError(problem)

	RespondString(ctx, writer, 500, "Internal error")
}

// ParseRequestEntity attempts to parse the request entity into the target entity.  True if successful, otherwise an
// error is written to the response and false is returned.
func ParseRequestEntity[E any](writer http.ResponseWriter, request *http.Request, entity *E) bool {
	requestEntity, err := io.ReadAll(request.Body)
	if err != nil {
		ClientError(writer, request, err)
		return false
	}

	if err := json.Unmarshal(requestEntity, entity); err != nil {
		UnprocessableEntity(request.Context(), writer, err.Error())
		return false
	}
	return true
}

func ClientError(writer http.ResponseWriter, request *http.Request, problem error) {
	ctx := request.Context()
	RespondString(ctx, writer, 400, "Client error")
}

func UnprocessableEntity(ctx context.Context, writer http.ResponseWriter, body string) {
	RespondString(ctx, writer, 422, body)
}

// Ok writes a 200 Ok response and streams serialized JSON out.
func Ok(writer http.ResponseWriter, request *http.Request, entity interface{}) {
	ctx := request.Context()
	//todo: encode to things other than JSON
	header := writer.Header()
	header.Add("Content-Type", "application/json")
	writer.WriteHeader(200)

	out := json.NewEncoder(writer)
	if err := out.Encode(entity); err != nil {
		span := trace.SpanFromContext(ctx)
		span.SetStatus(codes.Error, "failed to encode JSON")
		span.RecordError(err)
	}
}

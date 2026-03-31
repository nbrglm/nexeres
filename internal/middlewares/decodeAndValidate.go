package middlewares

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nbrglm/nexeres/internal/api/contracts"
	"github.com/nbrglm/nexeres/utils"
)

type CtxKeyBody struct{}

func DecodeAndValidate[T any]() func(nextHandler http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body T

			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				contracts.BadRequest("Invalid request body, error: " + err.Error()).Write(w)
				return
			}

			err = utils.Validator.Struct(body)
			if err != nil {
				contracts.BadRequest("Validation failed for request body, error: " + err.Error()).Write(w)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, CtxKeyBody{}, &body)
			r = r.WithContext(ctx)

			nextHandler.ServeHTTP(w, r)
		})
	}
}

// GetDecodedBody retrieves the decoded and validated body from the request context.
//
// If the body is not found or of the wrong type, this function will respond with a 400 Bad Request and return nil.
func GetDecodedBody[T any](ctx context.Context, w http.ResponseWriter) *T {
	body, ok := ctx.Value(CtxKeyBody{}).(*T)
	if !ok {
		contracts.BadRequest("Failed to retrieve decoded request body").Write(w)
		return nil
	}

	if body == nil {
		contracts.BadRequest("Decoded request body is not found").Write(w)
		return nil
	}

	return body
}

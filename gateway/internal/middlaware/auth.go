package middlaware

import (
	"context"
	"net/http"
	"strings"
	"time"

	pb "github.com/avran02/fileshare/proto/authpb"
)

const ContextUserIDKey = "userID"

func GetAuthMiddleware(authClient pb.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				http.Error(w, "Malformed token", http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			req := &pb.ValidateTokenRequest{AccessToken: token}
			resp, err := authClient.ValidateToken(ctx, req)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Store userID in context for future handlers
			ctx = context.WithValue(r.Context(), ContextUserIDKey, resp.Id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

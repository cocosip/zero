package cors

import (
	"github.com/gorilla/handlers"
	"net/http"
)

func Filter(opt *CorsOption) func(http.Handler) http.Handler {
	return FilterHandler(opt.GetOrigins(), opt.GetMethods(), opt.GetHeaders(), opt.GetAllowCredentials())
}

func FilterHandler(origins, methods, headers []string, allowCredentials bool) func(http.Handler) http.Handler {
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(headers) == 0 {
		headers = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With"}
	}

	var opts = []handlers.CORSOption{
		handlers.AllowedOrigins(origins),
		handlers.AllowedMethods(methods),
		handlers.AllowedHeaders(headers),
	}
	if allowCredentials {
		opts = append(opts, handlers.AllowCredentials())
	}
	return handlers.CORS(opts...)
}

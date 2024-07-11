package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time_now := time.Now().Format(time.DateTime)
		str := fmt.Sprintf("%s %s %s %s", r.Method, r.URL, r.Host, time_now)
		fmt.Println(str)

		next.ServeHTTP(w, r)
	})
}

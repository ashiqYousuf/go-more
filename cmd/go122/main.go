package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const AuthUserID = "go-more.auth.userID"

// ? Handlers
func pathParamHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.Write([]byte("Invalid ID supplied"))
		return
	}

	w.Write([]byte(fmt.Sprintf("ItemID %d", id)))
}

func methodBasedRouting(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(AuthUserID).(string)

	fmt.Fprintf(w, fmt.Sprintf("User with ID %v created item", userID))
}

func hostBasedRouting(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Host based routing")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Deleting an item")
}

func H(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("H"))
}

// ? Middlewares

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			// ? For Logging middleware next is CheckParams middleware
			// ? For Nth middleware (N+1)th will be the next one
			// ? For last middleware the next is servemux which forwards request to the correct handler
			next = x(next)
		}

		return next
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logging...")
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(r.Method, r.URL.Path, time.Since(start))
	})
}

func CheckPerms(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Checking perms...")
		next.ServeHTTP(w, r)
	})
}

func EnsureAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Sorry! Not an Admin!"))
		return
	})
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			w.Write([]byte("UnAuthorized"))
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			w.Write([]byte("UnAuthorized"))
			return
		}

		ctx := context.WithValue(r.Context(), AuthUserID, token)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

// ? Routes

func getRoutes() http.Handler {
	router := http.NewServeMux()

	stack := CreateStack(Logging, CheckPerms)

	router.HandleFunc("GET /item/{id}", pathParamHandler)
	router.Handle("POST /create-item", Auth(http.HandlerFunc(methodBasedRouting)))
	router.HandleFunc("domain.foo.bar/", hostBasedRouting)

	// ! Admin restricted routes
	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("DELETE /item/{id}", deleteItem)

	router.Handle("/", Auth(EnsureAdmin(adminRouter)))

	// ? sub-routes /v1/
	// v1 := http.NewServeMux()
	// v1.Handle("/v1/", http.StripPrefix("/v1", router))

	// return stack(v1)
	// ? End sub routes
	return stack(router)
}

func main() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: getRoutes(),
	}

	log.Println("Starting server on port :8080")
	server.ListenAndServe()
}

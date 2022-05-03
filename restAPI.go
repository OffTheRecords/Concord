package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func startRestAPI() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/auth/login", loginHandler)
	router.HandleFunc("/auth/register", registerHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Login")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Register")
}

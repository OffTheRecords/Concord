package main

import (
	"Concord/Authentication"
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"reflect"
)

type WebHandlerVars struct {
	dbClient *mongo.Database
}

func startRestAPI(dbClient *mongo.Database) {
	handlerVars := &WebHandlerVars{dbClient: dbClient}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/auth/login", handlerVars.loginHandler).Methods("POST")
	router.HandleFunc("/auth/register", handlerVars.registerHandler).Methods("POST", "OPTIONS")

	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:8080"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router)))
}

type login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type register struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type response struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func (vars *WebHandlerVars) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response response

	//Decode user response
	var login login
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		response.Status = 4004
		response.Msg = "Failed to decode response"

	}
	//TODO check all fields are set

	//TODO send to authentication server credentials for verification

	//TODO return auth token

	//Return success message
	response.Status = 200
	response.Msg = "ok"
	marshal, err := json.Marshal(response)
	if err != nil {
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		return
	}

	//DEBUG
	fmt.Printf("Got Login request with email: %s, passsword: %s\n", login.Email, login.Password)
}

func (vars *WebHandlerVars) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response response

	//Decode user response
	var register register
	err := json.NewDecoder(r.Body).Decode(&register)
	if err != nil {
		response.Status = 4005
		response.Msg = "Failed to decode response"
	} else {
		//check all fields are set
		ref := reflect.ValueOf(register)
		ty := ref.Type()
		fieldsErr := Authentication.FieldValidityCheck(ref, ty)
		if fieldsErr != nil {
			response.Status = fieldsErr.ErrorCode()
			response.Msg = fieldsErr.ErrorMsg()
		} else {
			//Fields okay
			//create new record in database for user
			gerr := Authentication.RegisterUser(register.Email, register.Username, register.Password, vars.dbClient)
			if gerr != nil {
				response.Status = gerr.ErrorCode()
				response.Msg = gerr.ErrorMsg()
			} else {
				response.Status = 200
				response.Msg = "ok"
			}
		}
	}

	//TODO return auth token for user

	//Return success message

	marshal, err := json.Marshal(response)
	if err != nil {
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		return
	}

	//DEBUG
	fmt.Printf("Got register request with email: %s, passsword: %s, username: %s\n", register.Email, register.Password, register.Username)

}

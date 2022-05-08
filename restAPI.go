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
	router.HandleFunc("/auth/login", handlerVars.loginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/register", handlerVars.registerHandler).Methods("POST", "OPTIONS")

	//TODO: ALLOW PASSING IN VARIABLE FOR ALLOWED ORIGIN
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:8080"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Access-Control-Allow-Headers", "access-control-allow-origin"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	allowedAuth := handlers.AllowCredentials()
	log.Fatal(http.ListenAndServe(":8081", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders, allowedAuth)(router)))
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

type user struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
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

	} else {
		//check all fields are set
		ref := reflect.ValueOf(login)
		fieldsErr := Authentication.FieldEmptyCheck(ref)
		if fieldsErr != nil {
			response.Status = fieldsErr.ErrorCode()
			response.Msg = fieldsErr.ErrorMsg()
		} else {
			jwt, err, expTime := Authentication.Login(login.Email, login.Password, vars.dbClient)
			if err != nil {
				fmt.Printf(jwt)
			} else {
				//fmt.Printf("jwt successful token: %s", jwt)
				//TODO GOTTA ADD SSL TO ALLOW SECURE COOKIES TO BE SAVED
				cookie := &http.Cookie{
					Name:     "token",
					Value:    jwt,
					Expires:  expTime,
					Path:     "/",
					Secure:   true,
					SameSite: 4,
				}
				http.SetCookie(w, cookie)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(200)
			}
		}
	}
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
	//fmt.Printf("Got Login request with email: %s, passsword: %s\n", login.Email, login.Password)
}

func (vars *WebHandlerVars) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
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
				//GENERATE AUTH TOKEN upon successful registration
				var userDetails user
				userDetails.Email = register.Email
				userDetails.Username = register.Username
				userDetails.Password = register.Password
				userDetails.Role = "Unverified"
				refUser := reflect.ValueOf(userDetails)
				jwt, err, expTime := Authentication.GenerateJWT(refUser)
				if err != nil {
					fmt.Print(err)
				} else {
					fmt.Printf("jwt successful token: %s", jwt)
					cookie := &http.Cookie{
						Name:     "token",
						Value:    jwt,
						Expires:  expTime,
						Path:     "/",
						Secure:   true,
						SameSite: 4,
					}
					http.SetCookie(w, cookie)
				}
				response.Status = 200
				response.Msg = "ok"
			}
		}
	}

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
	fmt.Printf("\nGot register request with email: %s, passsword: %s, username: %s\n", register.Email, register.Password, register.Username)

}

package main

import (
	"Concord/Authentication"
	"Concord/CustomErrors"
	"Concord/Structures"
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
	allowedOrigins := handlers.AllowedOrigins([]string{"https://localhost:8080", "https://127.0.0.1:8080"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Access-Control-Allow-Headers", "access-control-allow-origin", "Authorization"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	allowedAuth := handlers.AllowCredentials()
	ttl := handlers.MaxAge(60)
	err := http.ListenAndServeTLS(":443", "res/server_fullchain.pem", "res/server_privatekey.pem", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders, allowedAuth, ttl)(router))
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err)
	}
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

func (vars *WebHandlerVars) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response Structures.Response
	response.Status = 200
	response.Msg = "ok"

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
			jwt, gerr, expTime := Authentication.Login(login.Email, login.Password, vars.dbClient)
			if gerr != nil {
				CustomErrors.ErrorCodeHandler(gerr, &response)
			} else {
				//Set jwt token cookie
				cookie := &http.Cookie{
					Name:     "token",
					Value:    jwt,
					Expires:  expTime,
					MaxAge:   Authentication.JWT_TOKEN_TTL_MIN * 60,
					HttpOnly: true,
					Path:     "/",
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				}
				http.SetCookie(w, cookie)
			}
		}
	}

	//TODO Return refresh token

	//Return status message
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

	var response Structures.Response
	response.Status = 200
	response.Msg = "ok"

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
				CustomErrors.ErrorCodeHandler(gerr, &response)
			} else {
				user, gerr := Authentication.GetUserFromDB(register.Email, vars.dbClient)
				if gerr != nil {
					CustomErrors.ErrorCodeHandler(gerr, &response)
				} else {
					//GENERATE AUTH TOKEN upon successful registration
					jwt, err, expTime := Authentication.GenerateJWT(user)
					if err != nil {
						CustomErrors.LogError(5017, CustomErrors.LOG_WARNING, false, err)
						response.Status = 5017
						response.Msg = "internal server error"
					} else {
						cookie := &http.Cookie{
							Name:     "token",
							Value:    jwt,
							Expires:  expTime,
							Path:     "/",
							Secure:   true,
							SameSite: 4,
							HttpOnly: false,
							MaxAge:   900,
						}
						http.SetCookie(w, cookie)
					}
				}
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

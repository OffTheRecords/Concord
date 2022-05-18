package main

import (
	"Concord/Authentication"
	"Concord/CustomErrors"
	"Concord/Structures"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type WebHandlerVars struct {
	dbClient          *mongo.Database
	redisGlobalClient *redis.Client
}

func startRestAPI(dbClient *mongo.Database, redisGlobalClient *redis.Client) {
	handlerVars := &WebHandlerVars{dbClient: dbClient, redisGlobalClient: redisGlobalClient}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/auth/login", handlerVars.loginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/register", handlerVars.registerHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/refresh", handlerVars.refreshHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/user/{id:[0-9a-z]{24}$}", handlerVars.userGetHandler).Methods("GET", "OPTIONS")

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
			jwt, user, gerr := Authentication.Login(login.Email, login.Password, vars.dbClient)
			if gerr != nil {
				CustomErrors.GenericErrorCodeHandler(gerr, &response)
			} else {
				//Set jwt token cookie
				accessCookieExpiry := &http.Cookie{
					Name:     "accessTokenExpiry",
					Value:    strconv.FormatInt(jwt.AccessExpiry.Unix()-30, 10),
					MaxAge:   Authentication.JWT_TOKEN_TTL_MIN*60 - 30,
					Path:     "/",
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				}
				http.SetCookie(w, accessCookieExpiry)

				accessCookie := &http.Cookie{
					Name:     "accessToken",
					Value:    jwt.AccessToken,
					Expires:  jwt.AccessExpiry,
					MaxAge:   Authentication.JWT_TOKEN_TTL_MIN * 60,
					HttpOnly: true,
					Path:     "/",
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				}
				http.SetCookie(w, accessCookie)

				refreshCookie := &http.Cookie{
					Name:     "refreshToken",
					Value:    jwt.RefreshToken,
					Expires:  jwt.RefreshExpiry,
					MaxAge:   Authentication.REFRESH_TOKEN_TTL_MIN * 60,
					HttpOnly: true,
					Path:     "/auth/refresh",
					Secure:   true,
					SameSite: http.SameSiteNoneMode,
				}
				http.SetCookie(w, refreshCookie)

				//Json message response
				loginResponse := Structures.LoginResponse{ID: user.ID.Hex()}
				loginResponseJson, _ := json.Marshal(loginResponse)
				response.Msg = string(loginResponseJson)
			}
		}
	}

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
			_, gerr := Authentication.RegisterUser(register.Email, register.Username, register.Password, vars.dbClient)
			if gerr != nil {
				CustomErrors.GenericErrorCodeHandler(gerr, &response)
			} else {
				user, gerr := Authentication.GetUserFromDB(register.Email, vars.dbClient)
				if gerr != nil {
					CustomErrors.GenericErrorCodeHandler(gerr, &response)
				} else {
					//GENERATE AUTH TOKEN upon successful registration
					jwt, err := Authentication.GenerateJWT(user)
					if err != nil {
						CustomErrors.LogError(5017, CustomErrors.LOG_WARNING, false, err)
						response.Status = 5017
						response.Msg = "internal server error"
					} else {
						accessCookieExpiry := &http.Cookie{
							Name:     "accessTokenExpiry",
							Value:    strconv.FormatInt(jwt.AccessExpiry.Unix()-30, 10),
							MaxAge:   Authentication.JWT_TOKEN_TTL_MIN*60 - 30,
							Path:     "/",
							Secure:   true,
							SameSite: http.SameSiteNoneMode,
						}
						http.SetCookie(w, accessCookieExpiry)

						accessCookie := &http.Cookie{
							Name:     "accessToken",
							Value:    jwt.AccessToken,
							MaxAge:   Authentication.JWT_TOKEN_TTL_MIN * 60,
							HttpOnly: true,
							Path:     "/",
							Secure:   true,
							SameSite: http.SameSiteNoneMode,
						}
						http.SetCookie(w, accessCookie)

						refreshCookie := &http.Cookie{
							Name:     "refreshToken",
							Value:    jwt.RefreshToken,
							MaxAge:   Authentication.REFRESH_TOKEN_TTL_MIN * 60,
							HttpOnly: true,
							Path:     "/auth/refresh",
							Secure:   true,
							SameSite: http.SameSiteNoneMode,
						}
						http.SetCookie(w, refreshCookie)

						//Json message response
						registerResponse := Structures.RegisterResponse{ID: user.ID.Hex()}
						registerResponseJson, _ := json.Marshal(registerResponse)
						response.Msg = string(registerResponseJson)
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

func (vars *WebHandlerVars) refreshHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response Structures.Response
	response.Status = 200
	response.Msg = "ok"

	//check if cookie for refresh token set
	refreshToken, gerr := refreshTokenSetCheck(r)
	if gerr != nil {
		response.Status = gerr.ErrorCode()
		response.Msg = gerr.ErrorMsg()
		writeStatusMessage(w, &response)
		return
	}

	//check if refresh token is valid
	claim, gerr := Authentication.VerifyJWT(refreshToken)
	if gerr != nil {
		CustomErrors.GenericErrorCodeHandler(gerr, &response)
		writeStatusMessage(w, &response)
		return
	}

	//Determine key of access token in database
	redisKey := claim.ID.Hex() + ".rt." + strconv.FormatInt(claim.ExpiresAt, 10)

	//check if refresh token is not blacklisted
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	instances, err := vars.redisGlobalClient.Exists(ctx, redisKey).Result()
	if err != nil {
		CustomErrors.LogError(5020, CustomErrors.LOG_WARNING, false, err)
		CustomErrors.ErrorCodeHandler(5020, err, &response)
		writeStatusMessage(w, &response)
		return
	}
	if instances > 0 {
		gerr = CustomErrors.NewGenericError(4008, "refresh token blacklisted")
		CustomErrors.GenericErrorCodeHandler(gerr, &response)
		writeStatusMessage(w, &response)
		return
	}

	//Issue new jwt and refresh token
	user, gerr := Authentication.GetUserUsingIDFromDB(claim.ID.Hex(), vars.dbClient)
	if gerr != nil {
		CustomErrors.LogError(gerr.ErrorCode(), CustomErrors.LOG_WARNING, false, gerr)
		CustomErrors.GenericErrorCodeHandler(gerr, &response)
		writeStatusMessage(w, &response)
		return
	}
	jwt, err := Authentication.GenerateJWT(user)
	if err != nil {
		CustomErrors.LogError(5015, CustomErrors.LOG_WARNING, false, err)
		CustomErrors.ErrorCodeHandler(5015, err, &response)
		writeStatusMessage(w, &response)
		return
	}

	accessCookie := &http.Cookie{
		Name:     "accessToken",
		Value:    jwt.AccessToken,
		MaxAge:   Authentication.JWT_TOKEN_TTL_MIN * 60,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    jwt.RefreshToken,
		MaxAge:   Authentication.REFRESH_TOKEN_TTL_MIN * 60,
		HttpOnly: true,
		Path:     "/auth/refresh",
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, refreshCookie)

	//Add refresh token to blacklist
	redisKey = claim.ID.Hex() + ".rt." + strconv.FormatInt(jwt.RefreshExpiry.Unix(), 10)
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = vars.redisGlobalClient.Set(ctx, redisKey, jwt.RefreshToken, time.Duration(time.Minute*Authentication.REFRESH_TOKEN_TTL_MIN)).Err()
	if err != nil {
		CustomErrors.LogError(5022, CustomErrors.LOG_WARNING, false, err)
		CustomErrors.ErrorCodeHandler(5022, err, &response)
		writeStatusMessage(w, &response)
		return
	}

	//Return success message
	writeStatusMessage(w, &response)
}

func (vars *WebHandlerVars) userGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var response Structures.Response
	response.Status = 200
	response.Msg = "ok"

	accessToken, gerr := jwtSetCheck(r)
	if gerr != nil {
		response.Status = gerr.ErrorCode()
		response.Msg = gerr.ErrorMsg()
		writeStatusMessage(w, &response)
		return
	}

	claim, gerr := Authentication.VerifyJWT(accessToken)
	if gerr != nil {
		CustomErrors.GenericErrorCodeHandler(gerr, &response)
		writeStatusMessage(w, &response)
		return
	}

	if claim.ID.Hex() == mux.Vars(r)["id"] {
		user, gerr := Authentication.GetUserUsingIDFromDB(claim.ID.Hex(), vars.dbClient)
		if gerr != nil {
			CustomErrors.GenericErrorCodeHandler(gerr, &response)
			writeStatusMessage(w, &response)
			return
		}

		//Json message response
		userGetResponse := Structures.Users{ID: user.ID, Username: user.Username, Discriminator: user.Discriminator, Avatar: user.Avatar, Bot: user.Bot, System: user.System, MFA: user.MFA, Banner: user.Banner, Accent_Color: user.Accent_Color, Locale: user.Locale, Email: user.Email, EmailVerified: user.EmailVerified, Role: user.Role, RegisterTime: user.RegisterTime}
		userGetResponseJson, _ := json.Marshal(userGetResponse)
		response.Msg = string(userGetResponseJson)
		response.Status = 200
	} else {
		response.Status = 401

		response.Msg = "user " + claim.ID.Hex() + " unauthorized to view profile " + mux.Vars(r)["id"]
	}

	writeStatusMessage(w, &response)
}

func jwtSetCheck(r *http.Request) (string, CustomErrors.GenericErrors) {
	cookie, err := r.Cookie("accessToken")
	if err != nil {
		return "", CustomErrors.NewGenericError(4012, "accessToken cookie not found")
	}

	cookieValue := cookie.Value
	if len(cookieValue) == 0 {
		return "", CustomErrors.NewGenericError(4012, "accessToken cookie not found")
	}

	return cookieValue, nil
}

func refreshTokenSetCheck(r *http.Request) (string, CustomErrors.GenericErrors) {
	cookie, err := r.Cookie("refreshToken")
	if err != nil {
		return "", CustomErrors.NewGenericError(4014, "refreshToken cookie not found")
	}

	cookieValue := cookie.Value
	if len(cookieValue) == 0 {
		return "", CustomErrors.NewGenericError(4015, "refreshToken cookie not found")
	}

	return cookieValue, nil
}

func writeStatusMessage(w http.ResponseWriter, response *Structures.Response) {
	marshal, err := json.Marshal(response)
	if err != nil {
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		return
	}
}

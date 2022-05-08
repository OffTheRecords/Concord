package Authentication

import (
	"Concord/CustomErrors"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"time"
)

type Claims struct {
	Email      string `json:"email"`
	Role       string `json:"role"`
	Authorized bool   `json:"authorized"`
	jwt.StandardClaims
}

//Will check if the entered struct has any uninitialized types (either nil or empty string)
func FieldValidityCheck(value reflect.Value, ref reflect.Type) CustomErrors.GenericErrors {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return CustomErrors.NewGenericError(4006, "empty fields")
		}
		if ref.Field(i).Name == "Email" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4001, "invalid email format")
			}
		}
		if ref.Field(i).Name == "Username" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A[a-zA-Z0-9_-]{2,32}$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4002, "invalid username format")
			}

		}
		if ref.Field(i).Name == "Password" {
			val := value.Field(i).String()
			argMatched, err := regexp.MatchString("\\A.{8,128}$", val)
			if argMatched == false || err != nil {
				return CustomErrors.NewGenericError(4003, "invalid password format")
			}

		}
	}
	return nil
}

func FieldEmptyCheck(value reflect.Value) CustomErrors.GenericErrors {
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Interface() == nil || value.Field(i).Interface() == "" {
			return CustomErrors.NewGenericError(4006, "empty fields")
		}
	}
	return nil
}

func GenerateJWT(userDetails reflect.Value) (string, error, time.Time) {
	priv, err := ioutil.ReadFile("./private.pem")
	jwtKey, _ := jwt.ParseRSAPrivateKeyFromPEM(priv)

	expiration := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email:      userDetails.FieldByName("Email").String(),
		Role:       userDetails.FieldByName("Role").String(),
		Authorized: true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiration.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err, expiration
	}
	return tokenString, nil, expiration
}

func GeneratePrivateKey() {
	// generate key
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Cannot generate RSA key\n")
		os.Exit(1)
	}
	publickey := &privatekey.PublicKey

	// dump private key to file
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privatekey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	privatePem, err := os.Create("private.pem")
	if err != nil {
		fmt.Printf("error when create private.pem: %s \n", err)
		os.Exit(1)
	}
	err = pem.Encode(privatePem, privateKeyBlock)
	if err != nil {
		fmt.Printf("error when encode private pem: %s \n", err)
		os.Exit(1)
	}

	// dump public key to file
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		fmt.Printf("error when dumping publickey: %s \n", err)
		os.Exit(1)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	publicPem, err := os.Create("public.pem")
	if err != nil {
		fmt.Printf("error when create public.pem: %s \n", err)
		os.Exit(1)
	}
	err = pem.Encode(publicPem, publicKeyBlock)
	if err != nil {
		fmt.Printf("error when encode public pem: %s \n", err)
		os.Exit(1)
	}
}

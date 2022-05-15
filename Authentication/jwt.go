package Authentication

import (
	"Concord/CustomErrors"
	"Concord/Structures"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/ed25519"
	"io/ioutil"
	"os"
	"time"
)

const JWT_KEY_STORAGE_LOCATION = "res"
const JWT_TOKEN_TTL_MIN = 15
const REFRESH_TOKEN_TTL_MIN = 1440

var jwtPrivateKey crypto.PrivateKey
var jwtPublicKey crypto.PublicKey

type Claims struct {
	ID   primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Role int                `json:"roleid"`
	jwt.StandardClaims
}

//Generates server jwt tokens if they do not exist
func GenerateJWT(user Structures.Users) (Structures.AuthTokens, error) {

	//Add expiration times to structure
	expirationJWT := time.Now().Add(JWT_TOKEN_TTL_MIN * time.Minute)
	expirationRT := time.Now().Add(REFRESH_TOKEN_TTL_MIN * time.Minute)
	tk := Structures.AuthTokens{AccessExpiry: expirationJWT, RefreshExpiry: expirationRT}

	//Assign unique id to refresh token for revocation usage
	rt_uuid, err := uuid.NewRandom()
	if err != nil {
		return Structures.AuthTokens{}, err
	}
	tk.RefreshID = rt_uuid.String()

	//Load private key into memory so it does not need to be reloaded everytime its used
	if jwtPrivateKey == nil {
		priv, err := ioutil.ReadFile(JWT_KEY_STORAGE_LOCATION + "/jwt_private.pem")
		if err != nil {
			return tk, err
		}
		jwtPrivateKey, err = jwt.ParseEdPrivateKeyFromPEM(priv)
		if err != nil {
			return tk, err
		}
	}

	//Create jwt acccess token
	claims := &Claims{
		ID:   user.ID,
		Role: user.Role.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationJWT.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(jwtPrivateKey)
	if err != nil {
		return tk, err
	}
	tk.AccessToken = tokenString

	//Create Refresh Token
	claimsRefresh := Claims{
		ID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationRT.Unix(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claimsRefresh)
	refreshTokenString, err := refreshToken.SignedString(jwtPrivateKey)
	if err != nil {
		return tk, err
	}
	tk.RefreshToken = refreshTokenString

	return tk, nil
}

func GeneratePrivateKey() {
	//Create folder if it does not exist
	errG := makeRelativeDir(JWT_KEY_STORAGE_LOCATION)
	if errG != nil {
		CustomErrors.LogError(errG.ErrorCode(), CustomErrors.LOG_FATAL, true, errG)
	}

	// generate key
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		CustomErrors.LogError(5006, CustomErrors.LOG_FATAL, true, err)
	}

	//Dump private crypto keys to file
	cryptoPrivateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		CustomErrors.LogError(5008, CustomErrors.LOG_FATAL, true, err)
	}

	privateKeyBlock := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: cryptoPrivateKeyBytes,
	}

	privatePemWriter, err := os.OpenFile(JWT_KEY_STORAGE_LOCATION+"/jwt_private.pem", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		CustomErrors.LogError(5009, CustomErrors.LOG_FATAL, true, err)
	}

	err = pem.Encode(privatePemWriter, privateKeyBlock)
	if err != nil {
		CustomErrors.LogError(5010, CustomErrors.LOG_FATAL, true, err)
	}

	//Dump public crpyto key to file
	cryptoPublicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		CustomErrors.LogError(5011, CustomErrors.LOG_FATAL, true, err)
	}

	publicKeyBlock := &pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: cryptoPublicKeyBytes,
	}

	publicPemWriter, err := os.Create(JWT_KEY_STORAGE_LOCATION + "/jwt_public.pem")
	if err != nil {
		CustomErrors.LogError(5012, CustomErrors.LOG_FATAL, true, err)
	}
	err = pem.Encode(publicPemWriter, publicKeyBlock)
	if err != nil {
		CustomErrors.LogError(5013, CustomErrors.LOG_FATAL, true, err)
	}

}

func CheckAndCreateKeys() {
	_, err1 := os.Stat(JWT_KEY_STORAGE_LOCATION + "/jwt_private.pem")
	_, err2 := os.Stat(JWT_KEY_STORAGE_LOCATION + "/jwt_public.pem")
	if errors.Is(err1, os.ErrNotExist) || errors.Is(err2, os.ErrNotExist) {
		CustomErrors.LogError(2001, CustomErrors.LOG_INFO, false, errors.New("generating jwt keys"))
		GeneratePrivateKey()
	} else if err1 != nil && err2 != nil {
		CustomErrors.LogError(5007, CustomErrors.LOG_FATAL, true, errors.New("error checking if jwt keys exist"))
	} else {
		CustomErrors.LogError(2002, CustomErrors.LOG_INFO, false, errors.New("jwt keys already exist"))
	}
}

func VerifyJWT(accessToken string) (Claims, CustomErrors.GenericErrors) {

	// pass your custom claims to the parser function
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodEd25519)
		if !ok {
			return nil, errors.New("unexpected signing method")
		}

		//Load public key into memory so it does not need to be reloaded everytime its used
		if jwtPublicKey == nil {
			pub, err := ioutil.ReadFile(JWT_KEY_STORAGE_LOCATION + "/jwt_public.pem")
			if err != nil {
				return Claims{}, err
			}
			jwtPublicKey, err = jwt.ParseEdPublicKeyFromPEM(pub)
			if err != nil {
				return Claims{}, err
			}
		}
		return jwtPublicKey, nil
	})
	if err != nil {
		return Claims{}, CustomErrors.NewGenericError(5018, err.Error())
	}

	// type-assert `Claims` into a variable of the appropriate type
	myClaims := token.Claims.(*Claims)
	return *myClaims, nil
}

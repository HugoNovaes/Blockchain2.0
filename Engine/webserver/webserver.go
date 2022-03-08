package webserver

import (
	"encoding/json"
	"engine/utils"
	"engine/webserver/crud"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type (
	WebServer struct {
		port int
	}

	Login struct {
		Email string `json:"login"`
		Key   string `json:"key"`
	}

	JwtToken struct {
		Token string `json:"token"`
	}
)

var mySignature = []byte{0xa1, 0xae, 0x2a, 0xa1, 0x34, 0x68, 0x04, 0xce, 0xd2, 0xca, 0xa2, 0x95, 0x11, 0x29, 0x13, 0xea, 0x85, 0xc6, 0x9b, 0x8f}

func (w *WebServer) UsePort(port int) {
	w.port = port
}

func setHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.SetPrefix("\r")
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func doLogin(w http.ResponseWriter, r *http.Request) {

	var login Login
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		fmt.Printf("\r%s", err.Error())
	}

	claims := jwt.MapClaims{
		"email": login.Email,
		"key":   login.Key,
		"time":  time.Now(),
		"sign":  utils.EncodeHexString(mySignature),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, error := token.SignedString(mySignature)
	if error != nil {
		fmt.Println(error)
	}

	fmt.Printf("\r%v\r\n", login)

	json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

func createUser(w http.ResponseWriter, r *http.Request) {

	var user crud.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		fmt.Printf("\r%s", err.Error())
	}

	fmt.Printf("\r%v\r\n", user)

	newUser, err := user.Create()
	if err != nil {
		fmt.Printf("\r%s", err.Error())
		return
	}

	json.NewEncoder(w).Encode(newUser)
}

func (w *WebServer) Start(started chan bool) {
	r := mux.NewRouter()
	r.Use(setHeaders)

	r.HandleFunc("/login", doLogin).Methods("POST")
	r.HandleFunc("/newuser", createUser).Methods("POST")

	started <- true

	if w.port == 0 {
		w.port = 8080
	}

	tcpipAddress := fmt.Sprintf(":%d", w.port)

	log.Printf("Starting webserver on %s...\r\n", tcpipAddress)
	log.Fatal(http.ListenAndServe(tcpipAddress, r))
}

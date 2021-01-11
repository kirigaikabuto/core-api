package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/gorilla/mux"
	middleware "github.com/kirigaikabuto/common-lib/access-token-middleware"
	core_store "github.com/kirigaikabuto/core-store"
	"log"
	"net/http"
)

var (
	PORT string = "8000"
)

var cfgAmqp = amqp.Config{
	Host:        "localhost",
	VirtualHost: "",
	User:        "",
	Password:    "",
	Port:        5672,
	LogLevel:    5,
}

func main() {
	sess := amqp.NewSession(cfgAmqp)

	if err := sess.Connect(); err != nil {
		fmt.Println(err)
		return
	}
	defer sess.Close()

	var cltCfg = amqp.ClientConfig{
		//ResponseX: "response",
		//RequestX: "request",
	}
	clt, err := sess.Client(cltCfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	amqpClient := core_store.NewAmqpRequests(clt)

	var redisCfg = middleware.RedisConfig{
		Host: "127.0.0.1",
		Port: 6379,
	}
	accessTokenStore, err := middleware.NewAccessTokenStore(redisCfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	accessTokenMiddleware, err := middleware.NewAccessTokenMDW(accessTokenStore)
	if err != nil {
		fmt.Println(err)
		return
	}
	coreService := core_store.NewCoreService(*amqpClient, accessTokenStore)
	coreHttpEndpoints := core_store.NewHttpEndpoints(coreService)

	router := mux.NewRouter()
	router.Methods("GET").Path("/movies").Handler(accessTokenMiddleware.Middleware(http.HandlerFunc(coreHttpEndpoints.ListMoviesEndpoint())))
	router.Methods("GET").Path("/movies/{id}").Handler(accessTokenMiddleware.Middleware(http.HandlerFunc(coreHttpEndpoints.GetMovieByIdEndpoint("id"))))
	router.Methods("POST").Path("/register").Handler(http.HandlerFunc(coreHttpEndpoints.Register()))
	router.Methods("POST").Path("/login").Handler(http.HandlerFunc(coreHttpEndpoints.Login()))
	err = http.ListenAndServe("0.0.0.0:"+PORT, router)
	fmt.Println("running on port=", PORT)
	if err != nil {
		log.Fatal(err)
	}

}

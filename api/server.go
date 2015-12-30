package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/alexgear/gateway/config"
	"github.com/alexgear/gateway/gservices"
	"github.com/gorilla/mux"
)

var err error

//response structure to /duty
type GetDutyResponse struct {
	Number string `json:"number"`
}

func GetDutyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	cfg := config.GetConfig()
	number, err := gservices.GetDuty(cfg.CalendarId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GetDutyResponse{Number: number}
	toWrite, err := json.Marshal(response)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(toWrite)
	return
}

func InitServer(host string, port int) error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/duty", GetDutyHandler).Methods("GET")
	bind := fmt.Sprintf("%s:%d", host, port)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, router)
}

package main

import (
	"fmt"
	"utils"
	"net/http"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")
	config := utils.CargarConfiguracion[utils.ConfigMemory]("config.json")
	mux := http.NewServeMux()

	mux.HandleFunc("/interrupciones", utils.RecibirMensaje)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortMemory), mux)
	if err != nil {
		panic(err)
	}
}

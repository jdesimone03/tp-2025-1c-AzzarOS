package main

import (
	"fmt"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")

	config := utils.CargarConfiguracion[utils.ConfigMemory]("config.json")
	mux := http.NewServeMux()

	mux.HandleFunc("/peticiones", utils.RecibirMensaje)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortMemory), mux)
	if err != nil {
		panic(err)
	}
}

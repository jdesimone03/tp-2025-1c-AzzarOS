package main

import (
	"fmt"
	"net/http"
	"utils"
)

func main() {
	utils.ConfigurarLogger("log_MEMORIA")

	config := utils.CargarConfiguracion[utils.ConfigMemory]("config.json")


	http.HandleFunc("/peticiones", utils.RecibirInterfaz)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortMemory), nil)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"utils"
	"net/http"
)

func main() {
	utils.ConfigurarLogger("log_KERNEL")
	mux := http.NewServeMux()

	// mux.HandleFunc("/procesos", utils.RecibirPaquetes)
	mux.HandleFunc("/interrupciones", utils.RecibirMensaje)

	err := http.ListenAndServe(":8001", mux)
	if err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"net/http"
	"utils"
	"log/slog"
)

var interfazActual = utils.Interfaz{}

// TODO crear variable global que guarde la interfaz actual
func main() {
	utils.ConfigurarLogger("log_KERNEL")
	config := utils.CargarConfiguracion[utils.ConfigKernel]("config.json")

	http.HandleFunc("/interrupciones", recibirInterfaz)

	err := http.ListenAndServe(fmt.Sprintf(":%d",config.PortKernel), nil)
	if err != nil {
		panic(err)
	}
}

func recibirInterfaz(w http.ResponseWriter, r *http.Request) {
	interfaz, err := utils.DecodificarMensaje[utils.Interfaz](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	setInterfaz(interfaz)
	slog.Info(fmt.Sprintf("Me llego la siguiente interfaz: %+v",interfaz))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func setInterfaz(interfaz *utils.Interfaz){
	interfazActual = *interfaz
}
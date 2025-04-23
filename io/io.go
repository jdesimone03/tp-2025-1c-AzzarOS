package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"utils"
)

// TODO levantar servidor de IO despues del handshake
func main() {
	utils.ConfigurarLogger("log_IO")

	config := utils.CargarConfiguracion[utils.ConfigIO]("config.json")

	nombre := os.Args[1]

	interfaz := utils.Interfaz{
		Nombre: nombre,
		IP:     config.IPIo,
		Puerto: config.PortIo,
	}

	utils.EnviarMensaje(config.IPKernel, config.PortKernel, "handshakeIO", interfaz)

	http.HandleFunc("/peticionKernel", recibirPeticion)

	utils.IniciarServidor(config.PortIo)
}

func recibirPeticion(w http.ResponseWriter, r *http.Request) {
	peticion, err := utils.DecodificarMensaje[utils.PeticionKernel](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	// Log obligatorio 1
	slog.Info(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %d", peticion.PID, peticion.SuspensionTime))
	
	time.Sleep(time.Duration(peticion.SuspensionTime) * time.Millisecond)
	
	// Log obligatorio 2
	slog.Info(fmt.Sprintf("## PID: %d - Fin de IO", peticion.PID))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

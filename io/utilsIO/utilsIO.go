package utilsIO

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"utils"
	"utils/config"
	"utils/structs"
)

var Config = config.CargarConfiguracion[config.ConfigIO]("config.json")

func RecibirPeticion(w http.ResponseWriter, r *http.Request) {
	peticion, err := utils.DecodificarMensaje[structs.EsperaIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log obligatorio 1
	slog.Info(fmt.Sprintf("## PID: %d - Inicio de IO - Tiempo: %d", peticion.PID, peticion.TiempoMs))
	
	time.Sleep(time.Duration(peticion.TiempoMs) * time.Millisecond)
	
	// Log obligatorio 2
	slog.Info(fmt.Sprintf("## PID: %d - Fin de IO", peticion.PID))

	respuesta := structs.Respuesta{
		Mensaje: "IO_END",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
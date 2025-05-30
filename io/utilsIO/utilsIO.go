package utilsIO

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

var Config config.ConfigIO

func RecibirPeticion(w http.ResponseWriter, r *http.Request) {
	peticion, err := utils.DecodificarMensaje[structs.EjecucionIO](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log obligatorio 1/2
	logueador.InicioIO(peticion.PID, peticion.TiempoMs)

	time.Sleep(time.Duration(peticion.TiempoMs) * time.Millisecond)

	// Log obligatorio 2/2
	logueador.FinalizacionIO(peticion.PID)

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

package utilsIO

import (
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
	peticion, err := utils.DecodificarMensaje[structs.PeticionIO](r)
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
}

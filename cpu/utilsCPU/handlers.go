package utilsCPU

import (
	"net/http"
	"utils"
	"utils/logueador"
	"utils/structs"
	"encoding/json"		
)

// ---------------------------- Handlers ----------------------------//
func RecibirInterrupcion(w http.ResponseWriter, r *http.Request) {

	// Log obligatorio 2/11
	logueador.InterrupcionRecibida()
	InterruptFlag = true

	w.WriteHeader(http.StatusOK)
}

func RecibirEjecucion(w http.ResponseWriter, r *http.Request) {
	
	ejecucion, err := utils.DecodificarMensaje[structs.EjecucionCPU](r) // llega de kernel
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chEjecucion <- *ejecucion

	w.WriteHeader(http.StatusOK)
}

func MostrarCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(Cache)
	if err != nil {
		http.Error(w, "Error al serializar la cache", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func MostrarTLB(w http.ResponseWriter, r *http.Request) {
	logueador.Info("Mostrando contenido de la TLB")
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(TLB)
	if err != nil {
		http.Error(w, "Error al serializar la TLB", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

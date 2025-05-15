package utilsMemoria

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	// "strconv"
	"strings"
	"utils"
	"utils/config"
    "utils/structs"
)

var Config = config.CargarConfiguracion[config.ConfigMemory]("config.json")


func EjecutarArchivo(path string) []string {
    contenido, err := os.ReadFile(path)
    if err != nil {
        slog.Error("error leyendo el archivo de instrucciones")
		return nil
    }

    lineas := strings.Split(string(contenido), "\n")

    for _, linea := range lineas {
        linea = strings.TrimSpace(linea)
        if linea == "" {
            continue // ignorar líneas vacías
        }
    }
    return lineas
}

func BuscarLineaInstruccion(lineas []string,pc uint) string {
    return lineas[pc]
}

func EnviarInstruccion(w http.ResponseWriter, r *http.Request) {
    pc, err := utils.DecodificarMensaje[structs.PeticionMemoria](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

    lineas := EjecutarArchivo("instrucciones.txt") // Cambiar a la ruta correcta del archivo de instrucciones, hardcodeado por ahora
    linea := BuscarLineaInstruccion(lineas, pc.PC)

	respuesta := structs.Respuesta{
		Mensaje: linea,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}

func NuevoProceso(w http.ResponseWriter, r *http.Request) {
	_, err := utils.DecodificarMensaje[structs.Proceso](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO Implementar
	
	w.WriteHeader(http.StatusOK)
}

func CheckMemoria(w http.ResponseWriter, r *http.Request) {
	tam, err := utils.DecodificarMensaje[int](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ok := "false"
	if MemoriaDisponible(*tam) {
		ok = "true"
	}

	respuesta := structs.Respuesta {
		Mensaje: ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}


func MemoriaDisponible(MemoriaSolicitada int) bool{
	if Config.MemorySize >= MemoriaSolicitada {
		slog.Info(fmt.Sprintf("Memoria disponible: %d bytes", Config.MemorySize))
		return true
	}else {
		slog.Info(fmt.Sprintf("Memoria no disponible, me quedan: %d bytes", Config.MemorySize))
		return false
	}
}
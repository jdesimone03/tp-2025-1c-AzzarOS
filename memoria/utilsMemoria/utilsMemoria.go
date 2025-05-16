package utilsMemoria

import (
	"bufio"
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
var Procesos = make(map[uint][]string) // PID: lista de instrucciones

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

// Recibe un PID y PC, La memoria lo busca en sus procesos y lo devuelve.
func EnviarInstruccion(w http.ResponseWriter, r *http.Request) {
    proceso, err := utils.DecodificarMensaje[structs.PeticionMemoria](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

    //lineas := EjecutarArchivo("instrucciones.txt") // Cambiar a la ruta correcta del archivo de instrucciones, hardcodeado por ahora
    linea := Procesos[proceso.PID][proceso.PC]

	respuesta := structs.Respuesta{
		Mensaje: linea,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}

func NuevoProceso(w http.ResponseWriter, r *http.Request) {
	proceso, err := utils.DecodificarMensaje[structs.Proceso](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	archivo, err := os.Open(proceso.Instrucciones)
	if err != nil {
		slog.Error(fmt.Sprintf("(%d) No se pudo abrir el archivo %s.",proceso.PID, proceso.Instrucciones))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer archivo.Close()

	// Lee linea por linea
	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := scanner.Text()
		Procesos[proceso.PID] = append(Procesos[proceso.PID], linea)
	}

	if err := scanner.Err(); err != nil {
		slog.Error(fmt.Sprintf("(%d) No se pudo leer el archivo %s.",proceso.PID, proceso.Instrucciones))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func CheckMemoria(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tam, err := utils.DecodificarMensaje[int](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !MemoriaDisponible(*tam) {
		json.NewEncoder(w).Encode(structs.Respuesta{Mensaje:"No hay memoria disponible"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(structs.Respuesta{Mensaje:"OK"})
	w.WriteHeader(http.StatusOK)
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
package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	// "strconv"
	"strings"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
)

var Config config.ConfigMemory
var Procesos = make(map[uint][]string) // PID: lista de instrucciones

var EspacioUsuario []byte
var EspacioKernel [][][]byte

var Metricas = make(map[uint]structs.Metricas)

func IniciarEstructuras() {
	// Carga el espacio de usuario
	EspacioUsuario = make([]byte, Config.MemorySize)

	// Carga el espacio de kernel (Paginacion jerárquica multinivel)
	// Capaz despues lo cambio
	EspacioKernel = make([][][]byte, Config.NumberOfLevels)
	for i := range Config.NumberOfLevels {
		EspacioKernel[i] = make([][]byte, Config.EntriesPerPage)
		for j := range Config.EntriesPerPage {
			EspacioKernel[i][j] = make([]byte, Config.PageSize)
		}
	}
}

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
	proceso, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	linea := Procesos[proceso.PID][proceso.PC]

	// Log obligatorio 3/5
	logueador.ObtenerInstruccion(proceso.PID, proceso.PC, linea)

	respuesta := structs.Respuesta{
		Mensaje: linea,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}

func NuevoProceso(w http.ResponseWriter, r *http.Request) {
	proceso, err := utils.DecodificarMensaje[structs.NuevoProceso](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	archivo, err := os.Open(proceso.Instrucciones)
	if err != nil {
		slog.Error(fmt.Sprintf("(%d) No se pudo abrir el archivo %s.", proceso.PID, proceso.Instrucciones))
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
		slog.Error(fmt.Sprintf("(%d) No se pudo leer el archivo %s.", proceso.PID, proceso.Instrucciones))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log obligatorio 1/5
	logueador.MemoriaCreacionDeProceso(proceso.PID, proceso.Tamanio)

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
		json.NewEncoder(w).Encode(structs.Respuesta{Mensaje: "No hay memoria disponible"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(structs.Respuesta{Mensaje: "OK"})
	w.WriteHeader(http.StatusOK)
}

func MemoriaDisponible(MemoriaSolicitada int) bool {
	if Config.MemorySize >= MemoriaSolicitada {
		slog.Info(fmt.Sprintf("Memoria disponible: %d bytes", Config.MemorySize))
		return true
	} else {
		slog.Info(fmt.Sprintf("Memoria no disponible, me quedan: %d bytes", Config.MemorySize))
		return false
	}
}

// Operaciones
func Read(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawPID := r.URL.Query().Get("pid")
	pid, err := strconv.ParseUint(rawPID, 10, 32)

	read, err := utils.DecodificarMensaje[structs.ReadInstruction](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	leido := EspacioUsuario[read.Address:read.Size]

	logueador.LecturaEnEspacioDeUsuario(uint(pid), read.Address, read.Size)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(leido))
}

func Write(w http.ResponseWriter, r *http.Request) {

	rawPID := r.URL.Query().Get("pid")
	pid, err := strconv.ParseUint(rawPID, 10, 32)

	write, err := utils.DecodificarMensaje[structs.WriteInstruction](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	copy(EspacioUsuario[write.Address:], []byte(write.Data))

	logueador.EscrituraEnEspacioDeUsuario(uint(pid), write.Address, len(write.Data))

	w.WriteHeader(http.StatusOK)
}

func MoverProcesoASwap(w http.ResponseWriter, r *http.Request) {
	_, err := utils.DecodificarMensaje[uint](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	// Se responde que la operación fue registrada como exitosa, aunque no se modifique el estado interno.
	json.NewEncoder(w).Encode(structs.Respuesta{Mensaje: "OK"})
}

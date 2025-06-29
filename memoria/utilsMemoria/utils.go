package utilsMemoria

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"utils"
	"utils/config"
	"utils/logueador"
	"utils/structs"
	"fmt"
	"path/filepath"
	"time"
	"bufio"

)

var Config config.ConfigMemory

var Procesos = make(map[uint][]string) // PID: lista de instrucciones
var EspacioUsuario []byte // memoriaPrincipal
var Metricas = make(map[uint]structs.Metricas) // Metricas
var Ocupadas map[int]structs.FrameInfo // Ver de cambiar a un vector de PID nomas 
var TDPMultinivel map[uint]*structs.Tabla  

func IniciarEstructuras() {
	// Carga el espacio de usuario
	EspacioUsuario = make([]byte, Config.MemorySize)
	InicializarOcupadas()
	TDPMultinivel = make(map[uint]*structs.Tabla)
	CreacionArchivoSWAP()
}

func CantidadDeFrames() int {
	return Config.MemorySize / Config.PageSize
} 


func EjecutarArchivo(path string) []string {
	contenido, err := os.ReadFile(path)
	if err != nil {
		logueador.Error("error leyendo el archivo de instrucciones")
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

func CheckMemoria(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tam, err := utils.DecodificarMensaje[int](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
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
		logueador.Info("Memoria disponible: %d bytes", Config.MemorySize)
		return true
	} else {
		logueador.Info("Memoria no disponible, me quedan: %d bytes", Config.MemorySize)
		return false
	}
}

// // Operaciones
// func Read(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	rawPID := r.URL.Query().Get("pid")
// 	pid, err := strconv.ParseUint(rawPID, 10, 32)

// 	read, err := utils.DecodificarMensaje[structs.ReadInstruction](r)
// 	if err != nil {
// 		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	leido := EspacioUsuario[read.Address:read.Size]

// 	// Log obligatorio 4/5
// 	logueador.LecturaEnEspacioDeUsuario(uint(pid), read.Address, read.Size)

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(string(leido))
// }

// func Write(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	rawPID := r.URL.Query().Get("pid")
// 	pid, err := strconv.ParseUint(rawPID, 10, 32)

// 	write, err := utils.DecodificarMensaje[structs.WriteInstruction](r)
// 	if err != nil {
// 		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	copy(EspacioUsuario[write.Address:], []byte(write.Data))

// 	// Log obligatorio 4/5
// 	logueador.EscrituraEnEspacioDeUsuario(uint(pid), write.Address, len(write.Data))

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode("OK")
// }

func NoQuedanMasInstrucciones(pid uint, pc uint) bool {
	return pc >= uint(len(Procesos[pid]))
}

// ----------------------------------------- Metricas ----------------------------------------- //

func ExisteElPID(pid uint) bool {
	_, ok := Procesos[pid]
	return ok
}

func CrearMetricaDeProceso(pid uint) {
	metrica := structs.Metricas{}
	Metricas[pid] = metrica
}

func IncrementarMetricaEn(pid uint, campo string) {
	ok := ExisteElPID(pid)
	if !ok {
		logueador.Error("No existe el PID %d", pid)
		return
	}
	metrica := Metricas[pid]
	switch campo {
	case "AccesoATablas":
		metrica.AccesosATablas += uint(Config.NumberOfLevels)
	case "InstruccionesSolicitadas":
		metrica.InstruccionesSolicitadas++
	case "BajadasAlSWAP":
		metrica.BajadasASWAP++
	case "SubidasAmemoria":
		metrica.SubidasAMemoria++
	case "Lecturas":
		metrica.Lecturas++
	case "Escrituras":
		metrica.Escrituras++
	default:
		logueador.Error("Campo no valido: %s", campo)
		return
	}
	Metricas[pid] = metrica
	logueador.Info("Metrica: %s - Incrementada para el PID: %d - Valor actual: %+v",campo, pid, metrica)
}

func InformarMetricasDe(pid uint) {
	ok := ExisteElPID(pid)	
	if !ok {
		logueador.Info("No existe el PID %d", pid)
		return
	}
	metrica := Metricas[pid]
	logueador.DestruccionDeProceso(int(pid), metrica)
	delete(Metricas, pid)
}

// ---------------------------- Archivo Dump -------------------------------------- // 


func NombreDelArchivoDMP(pid string) string {
	return pid + "-" + time.Now().Format("2006-01-02-15-04-05") + ".dmp"
}

func CreacionArchivoDump(pid uint) (*os.File, error) {
	
	err := os.MkdirAll(Config.DumpPath, 0755)
	if err != nil {
    return nil, fmt.Errorf("error al crear el directorio de dumps: %w", err)
	}

	pathCorrectoDump := filepath.Base(Config.DumpPath)
	nombreArchivo := NombreDelArchivoDMP(strconv.Itoa(int(pid)))
	rutaCompleta := fmt.Sprintf("%s/%s", pathCorrectoDump, nombreArchivo)

	file, err := os.Create(rutaCompleta)
	if err != nil {
		return nil, fmt.Errorf("error al crear el archivo de dump: %w", err)
	}
	return file, nil
}

func EscribirDumpEnArchivo(file *os.File, pid uint, frames []string) error {
	contenido := []byte(strings.Join(frames, "\n"))
	_, err := file.Write(contenido)
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo de dump: %w", err)
	}
	logueador.Info("Dump del PID %d guardado exitosamente en %s", pid, file.Name())
	return nil
}


func CargarPIDconInstrucciones(path string, pid int) {
	instrucciones := LeerArchivoYGuardarInstrucciones(path)
	Procesos[uint(pid)] = instrucciones
}

func LeerArchivoYGuardarInstrucciones(path string) []string {
	var instrucciones []string
	file , err := os.Open(path)
	check("No se pudo abrir el archivo",err)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() { 
		instruccion := strings.TrimSpace(scanner.Text()) 
			if instruccion != ""{
			instrucciones = append(instrucciones, instruccion)
			}
	}
	if err := scanner.Err(); err != nil {
		check("Error al leer la instruccion",err)
	}
	defer file.Close() 
	return instrucciones // Devulve un ["JUMP 1", "ADD 2", "SUB 3"]
}

func check(mensaje string, e error) {
	if e != nil {
		logueador.Error(mensaje, "error", e)	
	}
}

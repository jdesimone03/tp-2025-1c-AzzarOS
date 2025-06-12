package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"net/http"
	"utils"
	"utils/logueador"
	"utils/structs"
	"os"
	"strconv"
)


// Recibe un PID y PC, La memoria lo busca en sus procesos y lo devuelve.
func HandlerFetch(w http.ResponseWriter, r *http.Request) {

	proceso, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
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
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	archivo, err := os.Open(proceso.Instrucciones)
	if err != nil {
		logueador.Error("(%d) No se pudo abrir el archivo %s.", proceso.PID, proceso.Instrucciones)
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
		logueador.Error("(%d) No se pudo leer el archivo %s.", proceso.PID, proceso.Instrucciones)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Log obligatorio 1/5
	logueador.MemoriaCreacionDeProceso(proceso.PID, proceso.Tamanio)

	w.WriteHeader(http.StatusOK)
}

func HandlerDeSuspension(w http.ResponseWriter, r *http.Request) {
	
	pid := r.URL.Query().Get("pid")
	if pid == "" {
		logueador.Error("PID no proporcionado")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logueador.Error("Error al convertir PID a entero")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ExisteElPID(uint(pidInt)) {
		logueador.Error("El PID %d no existe", pidInt)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	SwapInProceso(uint(pidInt)) 
	IncrementarMetricaEn(uint(pidInt), "BajadasAlSWAP") // Aumento la métrica de bajadas al SWAP del PID
	logueador.Info("Swapin del proceso con PID: %d", pidInt)
	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
}

func HandlerDeDesuspension(w http.ResponseWriter, r *http.Request) {
	
	tam := r.URL.Query().Get("tam")
	pid := r.URL.Query().Get("pid")
	tamInt, err := strconv.Atoi(tam)
	
	if err != nil {
		logueador.Error("Error al convertir el tamaño: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !HayEspacioParaInicializar(tamInt) {
		logueador.Error("No hay espacio para desuspender el proceso")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logueador.Error("Error al convertir PID a entero")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	SwapOutProceso(uint(pidInt))
	IncrementarMetricaEn(uint(pidInt), "SubidasAMemoria") // Aumento la métrica de subidas a memoria principal del PID
	
	logueador.Info("Swapout del proceso con PID: %s", pid)
	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
}

func HandlerDeFinalizacion(w http.ResponseWriter, r *http.Request) {
	
	pid := r.URL.Query().Get("pid")
	pidInt, err := strconv.Atoi(pid)
	LiberarMemoria(uint(pidInt))
	if err != nil {
		logueador.Error("Error al convertir PID a entero")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	InformarMetricasDe(uint(pidInt))
	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
}

func HandlerWrite(w http.ResponseWriter, r *http.Request) {

	rawPID := r.URL.Query().Get("pid")
	pid, err := strconv.ParseUint(rawPID, 10, 32)

	write, err := utils.DecodificarMensaje[structs.WriteInstruction](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Write(uint(pid),write.Address, write.Data) // Aquí deberías convertir los datos a bytes antes de escribir	

	// Log obligatorio 4/5
	logueador.EscrituraEnEspacioDeUsuario(uint(pid), write.Address, len(write.Data))

	w.WriteHeader(http.StatusOK)
}

func HandlerRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rawPID := r.URL.Query().Get("pid")
	pid, err := strconv.ParseUint(rawPID, 10, 32)

	read, err := utils.DecodificarMensaje[structs.ReadInstruction](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	valorLeido, err := Read(uint(pid), read.Address, read.Size) // Aquí deberías convertir los datos a bytes antes de escribir
	if err != nil {
		logueador.Error("Error al leer de memoria: %e", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Log obligatorio 4/5
	logueador.LecturaEnEspacioDeUsuario(uint(pid), read.Address, read.Size)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(valorLeido))
}

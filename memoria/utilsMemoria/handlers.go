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
	"strings"
)

func HandlerPedidoTDP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	pid := r.URL.Query().Get("pid")
	if pid == "" {
		http.Error(w, "PID no proporcionado", http.StatusBadRequest)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logueador.Error("Error al convertir PID a entero: %v", err)
		http.Error(w, "Error al convertir PID a entero", http.StatusBadRequest)
		return
	}

	tablaDeProceso := TDPMultinivel[uint(pidInt)]
	if tablaDeProceso == nil {
		logueador.Error("No se encontró la tabla de páginas para el PID: %d", pidInt)
		http.Error(w, "Tabla de páginas no encontrada", http.StatusNotFound)
		return
	}

	tablaJSON, err := json.Marshal(tablaDeProceso)
	if err != nil {
		logueador.Error("Error al convertir la tabla de páginas a JSON: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(tablaJSON)
	logueador.Info("Tabla de páginas enviada para el PID: %d", pidInt)	
}

func HandlerConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	config := structs.ConfigMemoria{
		CantNiveles:         Config.NumberOfLevels,
		EntradasPorTabla:       Config.EntriesPerPage,
		TamanioPagina:   Config.PageSize,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		logueador.Error("Error al convertir la configuración a JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(configJSON)
	logueador.Info("Configuración enviada")
}

func HandlerCache(w http.ResponseWriter, r *http.Request) {
	paginaJSON, err := utils.DecodificarMensaje[structs.PaginaCache](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	contenidoStr := string(paginaJSON.Contenido)
	Write(uint(paginaJSON.PID), paginaJSON.NumeroFrame * Config.PageSize, contenidoStr ) 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}


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
	if err != nil {
		logueador.Error("Error al convertir PID a entero: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	if err != nil {
		logueador.Error("Error al convertir PID a entero: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

func HandlerMEMORYDUMP(w http.ResponseWriter, r *http.Request) {
	pid := r.URL.Query().Get("pid")
	if pid == "" {
		http.Error(w, "PID no proporcionado", http.StatusBadRequest)
		return
	}
	pidUint, err := strconv.ParseUint(pid, 10, 32)
	if err != nil {
		http.Error(w, "Error al convertir PID a entero", http.StatusBadRequest)
		return
	}
	file, err := CreacionArchivoDump(uint(pidUint))  
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError) // En el handler del kernel, si el httpStatus es este, manda el proceso a EXIT
		logueador.Info("Error al crear el archivo de dump de memoria: %v", err)
		return
	}
	defer file.Close() 
	logueador.MemoryDump(uint(pidUint)) 
	w.WriteHeader(http.StatusOK) // Salió todo bien
}

func HandlerPedidoDeInstruccion(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var data structs.EjecucionCPU 
		err := decoder.Decode(&data)
		logueador.Info("Error decodificando el cuerpo: %v", err)

		// Primer chequeo por si el PID no existe
		if !ExisteElPID(data.PID) {
			logueador.Info("No existe el PID: %d", data.PID)
			http.Error(w, "PID no existe", http.StatusBadRequest)
			return
		}

		// Segundo chequeo por si el PC es mayor al tamaño de las instrucciones => ya no quedan más instrucciones por ejecutar 
		if NoQuedanMasInstrucciones(data.PID, data.PC) {
			logueador.Info("No quedan más instrucciones para el PID: %d", data.PID)
			http.Error(w, "No quedan más instrucciones", http.StatusBadRequest)
			return
		}

		instruccion := Procesos[data.PID][data.PC] // Obtengo la instrucción del PID y el PC. 
		logueador.ObtenerInstruccion(data.PID, data.PC, instruccion) 
		MandarInstruccion(instruccion,w, r) // Envio la instruccion a la CPU
		IncrementarMetricaEn(data.PID, "InstruccionesSolicitadas") // Aumento la métrica de instrucciones solicitadas del PID
		
		w.WriteHeader(http.StatusOK) 
		w.Write([]byte("OK"))
}

func MandarInstruccion(instruccion string, w http.ResponseWriter, r *http.Request) {
	instruccionJSON, err := json.Marshal(instruccion)
	if err != nil {
		logueador.Info("Error al convertir la instrucción a JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(instruccionJSON)
	logueador.Info("Instrucción enviada desde memoria - instruccion: %s", instruccion)
}

func check(mensaje string, e error) {
	if e != nil {
		logueador.Error(mensaje, "error", e)	
	}
}

func HandlerDePedidoDeInicializacion(w http.ResponseWriter, r *http.Request) {
		
		decoder := json.NewDecoder(r.Body)
		var data structs.PedidoDeInicializacion
		err := decoder.Decode(&data)
		check("Error decodificando", err)
		if ExisteElPID(data.PID) {
			logueador.Info("El PID ya existe: %d", data.PID)
			http.Error(w, "El PID ya existe", http.StatusBadRequest)
			return 
		}

		tamanioInt := int(data.TamanioProceso)

		if HayEspacioParaInicializar(tamanioInt) {
			logueador.Info("Hay espacio para el proceso con PID: %d", data.PID)
			
			CargarPIDconInstrucciones(data.Path, int(data.PID))  // Carga las instrucciones del PID en el map
			CrearMetricaDeProceso(data.PID) // Crea la metrica del proceso para ir guardando registro de las acciones
			CrearTablaDePaginas(data.PID, int(data.TamanioProceso)) // Crea la tabla de paginas del PID
			
			w.WriteHeader(http.StatusOK) // Envio el OK al kernel
			w.Write([]byte("OK")) // Envio el OK al kernel
			return  
		} else {
			logueador.Info("No hay espacio para el proceso con PID: %d", data.PID)
			http.Error(w, "No hay espacio para el proceso", http.StatusBadRequest) // Envio el error al kernel
			return
		}
	}


func CargarPIDconInstrucciones(path string, pid int) {
	instrucciones := LeerArchivoYGuardarInstrucciones(path)
	Procesos[uint(pid)] = instrucciones
	logueador.Info("PID: %d cargado con sus instrucciones: %s", pid, strings.Join(instrucciones, "-"))
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

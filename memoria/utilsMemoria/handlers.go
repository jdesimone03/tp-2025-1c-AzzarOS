package utilsMemoria

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"
	"utils"
	"utils/logueador"
	"utils/structs"
)

func HandlerHayEspacio(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	tam, err := utils.DecodificarMensaje[int](r)
	if err != nil {
		logueador.Error("Error al decodificar el mensaje: %v", err)
		http.Error(w, "Error al decodificar el mensaje", http.StatusBadRequest)
		return
	}

	if HayEspacioParaInicializar(*tam) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hay espacio disponible"))
		logueador.Info("Hay espacio disponible para %d bytes", tam)
	} else {
		w.WriteHeader(http.StatusNoContent)
		logueador.Info("No hay espacio disponible para %d bytes", tam)
	}
}

func HandlerPedidoFrame(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	pid := r.URL.Query().Get("pid")
	direccion := r.URL.Query().Get("direccion")
	if pid == "" || direccion == "" {
		http.Error(w, "Datos no proporcionados", http.StatusBadRequest)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logueador.Error("Error al convertir PID a entero: %v", err)
		http.Error(w, "Error al convertir PID a entero", http.StatusBadRequest)
		return
	}

	if !ExisteElPID(uint(pidInt)) {
		logueador.Error("El PID %s no existe", pid)
		http.Error(w, "PID no existe", http.StatusBadRequest)
		return
	}

	direccionInt, err := strconv.Atoi(direccion)
	if err != nil {
		logueador.Error("Error al convertir dirección a entero: %v", err)
		http.Error(w, "Error al convertir dirección a entero", http.StatusBadRequest)
		return
	}

	frame := direccionInt / Config.PageSize
	inicio := frame * Config.PageSize
	fin := inicio + Config.PageSize
	paginaADar := EspacioUsuario[inicio:fin] // Obtengo la pagina que corresponde al frame

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(paginaADar); err != nil {
		logueador.Error("Error al codificar página: %v", err)
	}
	logueador.Info("Página enviada para el PID: %s, Dirección: %s", pid, direccion)
}

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

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

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
	IncrementarMetricaEn(uint(pidInt), "AccesoATablas") // Aumento la métrica de tablas de páginas solicitadas del PID
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
		CantNiveles:      Config.NumberOfLevels,
		EntradasPorTabla: Config.EntriesPerPage,
		TamanioPagina:    Config.PageSize,
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		logueador.Error("Error al convertir la configuración a JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	w.WriteHeader(http.StatusOK)
	w.Write(configJSON)
	logueador.Info("Configuración enviada")
}

func HandlerEscribirDeCache(w http.ResponseWriter, r *http.Request) {


	paginaJSON, err := utils.DecodificarMensaje[structs.PaginaCache](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	contenidoStr := string(paginaJSON.Contenido)
	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio
	Write(uint(paginaJSON.PID), paginaJSON.NumeroFrame*Config.PageSize, contenidoStr)

	logueador.Info("Se escribió correctamente en memoria el frame %d del PID %d", paginaJSON.NumeroFrame, paginaJSON.PID)
	logueador.Info("[MEMORIA] Página actualizada: PID %d, Página %d, Contenido: %s\n", paginaJSON.PID, paginaJSON.NumeroPagina, paginaJSON.Contenido)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK")) // Envio un OK al kernel
}

func HandlerFetch(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	proceso, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	linea := Procesos[proceso.PID][proceso.PC]

	// Log obligatorio 3/5
	logueador.ObtenerInstruccion(proceso.PID, proceso.PC, linea)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(linea))
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

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	pid, err := utils.DecodificarMensaje[uint](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje (%v)", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ExisteElPID(*pid) {
		logueador.Error("El PID %d no existe", *pid)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logueador.Info("Existe el PID")
	time.Sleep(time.Millisecond * time.Duration(Config.SwapDelay)) // Simula el tiempo de espera para la verificación de espacio

	SwapInProceso(*pid)
	IncrementarMetricaEn(*pid, "BajadasAlSWAP") // Aumento la métrica de bajadas al SWAP del PID

	logueador.Info("SwapIn del proceso con PID: %d", *pid)

	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
	w.Write([]byte("OK"))
}

func HandlerDeDesuspension(w http.ResponseWriter, r *http.Request) {

	tam := r.URL.Query().Get("tam")
	pid := r.URL.Query().Get("pid")
	tamInt, err := strconv.Atoi(tam)

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	if err != nil {
		logueador.Error("Error al convertir el tamaño: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !HayEspacioParaInicializar(tamInt) {
		logueador.Error("No hay espacio para desuspender el proceso")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No hay espacio para desuspender el proceso"))
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

	time.Sleep(time.Millisecond * time.Duration(Config.SwapDelay)) // Simula el tiempo de espera para la verificación de espacio
	SwapOutProceso(uint(pidInt))
	IncrementarMetricaEn(uint(pidInt), "SubidasAmemoria") // Aumento la métrica de subidas a memoria principal del PID

	logueador.Info("Swapout del proceso con PID: %s", pid)
	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
}

func HandlerDeFinalizacion(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	pid, err := utils.DecodificarMensaje[uint](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ExisteElPID(*pid) {
		logueador.Error("El PID %d no existe", *pid)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	InformarMetricasDe(*pid)
	LiberarMemoria(*pid)

	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
	w.Write([]byte("OK"))
}

func HandlerWrite(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	write, err := utils.DecodificarMensaje[structs.WriteInstruction](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !ProcesoEnMemoria(write.PID) {
		logueador.Error("El proceso %d no está en memoria para poder escribir", write.PID)
		return
	}

	err = Write(write.PID, write.LogicAddress, write.Data) // Aquí deberías convertir los datos a bytes antes de escribir
	if err != nil {
		logueador.Error("Error al escribir en memoria: %e", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	IncrementarMetricaEn(write.PID, "Escrituras") // Aumenta la métrica de escrituras de memoria del PID
	// Log obligatorio 4/5
	logueador.EscrituraEnEspacioDeUsuario(write.PID, write.LogicAddress, len(write.Data))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func HandlerRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	read, err := utils.DecodificarMensaje[structs.ReadInstruction](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio
	valorLeido, err := Read(read.PID, read.Address, read.Size) // Aquí deberías convertir los datos a bytes antes de escribir
	if err != nil {
		logueador.Error("Error al leer de memoria: %e", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ProcesoEnMemoria(read.PID) {
		logueador.Error("El proceso %d no está en memoria. No se puede leer - ", read.PID)
		return
	}
	// Log obligatorio 4/5
	logueador.LecturaEnEspacioDeUsuario(read.PID, read.Address, read.Size)
	IncrementarMetricaEn(read.PID, "Lecturas") // Aumenta la métrica de lecturas de memoria del PID
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(valorLeido))
}

func HandlerMEMORYDUMP(w http.ResponseWriter, r *http.Request) {


	pid, err := utils.DecodificarMensaje[uint](r)
	if err != nil {
		logueador.Error("No se pudo decodificar el mensaje: %e", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio
	file, err := CreacionArchivoDump(*pid)
	if err != nil {
		logueador.Error("Error al crear el archivo de dump de memoria: %v", err)
		w.WriteHeader(http.StatusInternalServerError) // En el handler del kernel, si el httpStatus es este, manda el proceso a EXIT
		return
	}
	paginasDeProceso := BuscarPaginasDeProceso(*pid)    // Obtengo las paginas del proceso
	EscribirDumpEnArchivo(file, *pid, paginasDeProceso) // Escribo las paginas en el archivo de dump

	defer file.Close()
	logueador.MemoryDump(*pid)
	w.WriteHeader(http.StatusOK) // Salió todo bien
	w.Write([]byte("OK"))
}

func HandlerPedidoDeInstruccion(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio
	
	data, err := utils.DecodificarMensaje[structs.EjecucionCPU](r)
	if err != nil {
		logueador.Error("Error decodificando el cuerpo: %v", err)
		http.Error(w, "Error decodificando el cuerpo", http.StatusBadRequest)
	}

	// Primer chequeo por si el PID no existe

	if !ExisteElPID(data.PID) {
		logueador.Error("No existe el PID: %d", data.PID)
		http.Error(w, "PID no existe", http.StatusBadRequest)
		return
	}

	// Segundo chequeo por si el PC es mayor al tamaño de las instrucciones => ya no quedan más instrucciones por ejecutar
	if NoQuedanMasInstrucciones(data.PID, data.PC) {
		logueador.Error("No quedan más instrucciones para el PID: %d", data.PID)
		http.Error(w, "No quedan más instrucciones", http.StatusBadRequest)
		return
	}

	instruccion := Procesos[data.PID][data.PC] // Obtengo la instrucción del PID y el PC.
	logueador.ObtenerInstruccion(data.PID, data.PC, instruccion)
	// MandarInstruccion(instruccion, w, r)                       // Envio la instruccion a la CPU
	IncrementarMetricaEn(data.PID, "InstruccionesSolicitadas") // Aumento la métrica de instrucciones solicitadas del PID
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(instruccion))
}

func MandarInstruccion(instruccion string, w http.ResponseWriter, r *http.Request) {
	instruccionJSON, err := json.Marshal(instruccion)
	if err != nil {
		logueador.Error("Error al convertir la instrucción a JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(instruccionJSON)
	logueador.Info("Instrucción enviada desde memoria - instruccion: %s", instruccion)
}

func HandlerDePedidoDeInicializacion(w http.ResponseWriter, r *http.Request) {

	time.Sleep(time.Millisecond * time.Duration(Config.MemoryDelay)) // Simula el tiempo de espera para la verificación de espacio

	data, err := utils.DecodificarMensaje[structs.NuevoProceso](r)
	if err != nil {
		logueador.Error("Error al decodificar el mensaje: %v", err)
		http.Error(w, "Error al decodificar el mensaje", http.StatusBadRequest)
		return
	}

	if ExisteElPID(data.PID) {
		logueador.Error("El PID ya existe: %d", data.PID)
		http.Error(w, "El PID ya existe", http.StatusBadRequest)
		return
	}

	logueador.Info("Hay esapcio para inicializar el proceso con PID: %d", data.PID)

	// Creación de Estructuras
	CargarPIDconInstrucciones(data.Instrucciones, int(data.PID))     // Carga las instrucciones del PID en el map
	CrearMetricaDeProceso(data.PID)                         // Crea la metrica del proceso para ir guardando registro de las acciones
	CrearTablaDePaginas(data.PID, int(data.Tamanio)) // Crea la tabla de paginas del PID
	logueador.Info("Se han creado todas las estructuras necesarias para el PID: %d", data.PID)

	// Log obligatorio 1/5
	logueador.MemoriaCreacionDeProceso(data.PID, data.Tamanio)

	w.WriteHeader(http.StatusOK) // Envio el OK al kernel
	w.Write([]byte("OK"))        // Envio el OK al kernel
}

package utilsCPU

import (
	"encoding/json"
	"fmt"
	"net/http"
	"utils"
	"utils/logueador"
	"utils/structs"
	"strconv"
)

// --------------------------------------- Cache ---------------------------------------------

/*
Cosas para testear:
- Agregado de pagina a cache => Testeado
- Remplazo de pagina en cache:
  - Con Algoritmo CLOCK 
  - Con CLOCK - M 
  - Con FIFO => en TLB Testeado
  - Con LRU 
- Verificacion de si una pagina fue modificada => Testeado
- Envio de pagina a memoria
*/
var Cache structs.CacheStruct

func InicializarCache() {
	paginas := make([]structs.PaginaCache, Config.CacheEntries) // Slice vacío, capacidad predefinida
	
		for i := 0; i < Config.CacheEntries; i++ {
		paginas[i] = structs.PaginaCache{
			NumeroPagina: -1,
			NumeroFrame: -1,
			PID: -1,
			}
		}
	Cache = structs.CacheStruct{
		Paginas: paginas,
		Algoritmo: Config.CacheReplacement,
	}
}

func CacheHabilitado() bool {
	return Config.CacheEntries > 0 
}

func FueModificada(pagina structs.PaginaCache) bool {
	return pagina.BitModificado
}

func EstaEnCache(pid uint, direccionLogica int) bool {
	if !CacheHabilitado() {
		logueador.Info("Caché no habilitada, no se puede verificar si la página está en caché")
		return false 
	}

	paginaLogica := direccionLogica / ConfigMemoria.TamanioPagina // Obtenemos el número de página

	for _, pagina := range Cache.Paginas {
		if pagina.PID == int(pid) && pagina.NumeroPagina == paginaLogica && pagina.BitPresencia {
			return true // La página está en la caché
		}
	}
	return false 
}


func ObtenerPaginaDeCache(pid uint, nropagina int) int {
	
	if !CacheHabilitado() {
		logueador.Info("Caché no habilitada, no se puede obtener la página de caché")
		return -1
	}

	for i, pagina := range Cache.Paginas {
		if pagina.PID == int(pid) && pagina.NumeroPagina == nropagina && pagina.BitPresencia {
			logueador.Info("Página encontrada en caché: PID %d, Página %d", pid, nropagina)
			return i // Retorna la página y su índice en caché
		}
	}
	return -1
}

func MandarDatosAMP(paginas structs.PaginaCache) {
	
	respuesta := utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "actualizarMP", paginas)
	if respuesta != "OK" {
		logueador.Error("Error al enviar la pagina a memoria: %s", respuesta)
		return
	}
	logueador.Info("Pagina enviada a la memoria correctamente")
}

func PaginasModificadasCache(pid uint) []structs.PaginaCache {
	var paginasModificadas []structs.PaginaCache
	for _, pagina := range Cache.Paginas {
		if FueModificada(pagina) && pagina.PID == int(pid) { // Verificamos si la pagina fue modificada y pertenece al PID
			paginasModificadas = append(paginasModificadas, pagina)
		}
	}
	return paginasModificadas
}

// Debe venir una request de memoria o kernel
func DesaolojoDeProceso(w http.ResponseWriter, r *http.Request){
	
	pidStr := r.URL.Query().Get("pid") // Obtenemos el PID del proceso a desalojar
	pidInt, err := strconv.Atoi(pidStr)
	if err != nil {
		logueador.Info("Error al convertir el PID a entero: %v", err)
		http.Error(w, "PID inválido", http.StatusBadRequest)
		return
	}
	modificadas := PaginasModificadasCache(uint(pidInt))
	if len(modificadas) == 0 {
		w.Write([]byte("No hay paginas modificadas. No se actualiza memoria"))
		w.WriteHeader(http.StatusOK) // No hay paginas modificadas, todo bien
		return 
	}
	DesalojarCache(modificadas)
	DesalojoTlB(uint(pidInt)) // Desalojamos las entradas de TLB del PID
	EliminarEntradasDeCache(uint(pidInt)) // Eliminamos las entradas de cache del PID
	return 
}

func DesalojarCache(modificadas []structs.PaginaCache) {
	for i:=0; i < len(modificadas); i++ {
		contenido := modificadas[i].Contenido // Obtenemos el contenido de la pagina modificada
		utils.EnviarMensaje(Config.IPMemory, Config.PortMemory, "actualizarMP", contenido) // Enviamos el contenido a memoria)	
	}
	return
}


func EliminarEntradasDeCache(pid uint) {
	logueador.Info("Eliminando entradas de caché para el PID %d", pid)
	for i := 0; i < len(Cache.Paginas); i++ {
		if Cache.Paginas[i].PID == int(pid) { // Si el PID coincide, eliminamos la entrada
			Cache.Paginas[i] = structs.PaginaCache{} // Reinicializamos la entrada
			logueador.Info("Entrada de caché eliminada para el PID %d", pid)
		}
	}
}

func CreacionDePaginaCache(pid uint, nropagina int, contenido []byte, frame int) structs.PaginaCache {
	return structs.PaginaCache{
		NumeroPagina: nropagina,
		NumeroFrame: frame, // Asignamos el numero de frame
		BitPresencia: true, // La pagina esta presente en memoria
		BitModificado: false, // Inicialmente no ha sido modificada
		BitDeUso: true, // Inicialmente se considera que la pagina ha sido usada
		PID: int(pid), // Asignamos el PID del proceso
		Contenido: contenido, // Asignamos el contenido de la pagina
	}
}

func PedirFrameAMemoria(pid uint, direccionLogica int, direccionFisica int) (structs.PaginaCache, error) {

	if !CacheHabilitado() {
		logueador.Error("Cache no habilitada, no se puede pedir frame a memoria")
		return structs.PaginaCache{}, fmt.Errorf("cache no habilitada, no se puede pedir frame a memoria")
	}
	
	logueador.Info("Pidiendo frame a memoria para PID %d, direccion logica %d, direccion fisica %d", pid, direccionLogica, direccionFisica)
	nropagina := direccionLogica / ConfigMemoria.TamanioPagina // Obtenemos el numero de pagina
	url := fmt.Sprintf("http://%s:%s/pedirFrame?pid=%d&direccion=%d", Config.IPMemory, Config.PortMemory, pid, direccionFisica)
	
	resp, err := http.Get(url)
	if err != nil {
		logueador.Error("Error al pedir el frame a memoria: %v", err)
		return structs.PaginaCache{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		logueador.Error("Error al pedir el frame a memoria, status code: %d", resp.StatusCode)
		return structs.PaginaCache{}, fmt.Errorf("error al pedir el frame a memoria, status code: %d", resp.StatusCode)
	}

	var frame []byte 
	err = json.NewDecoder(resp.Body).Decode(&frame)
	if err != nil {
		logueador.Error("Error al decodificar el frame: %v", err)
		return structs.PaginaCache{}, err
	}

	logueador.Info("Frame recibido de memoria")
	paginaCache := CreacionDePaginaCache(pid, nropagina, frame, direccionFisica / ConfigMemoria.TamanioPagina) // Creamos la pagina cache con el frame obtenido

	return paginaCache, nil
}

func CacheLleno() bool {
	return len(Cache.Paginas) == Config.CacheEntries // Verificamos si la cantidad de paginas en cache es igual a la cantidad de entradas configuradas
}

func IndiceLibreCache() int {
	for i := 0; i < len(Cache.Paginas); i++ {
		if Cache.Paginas[i].NumeroPagina == -1 { // Si hay una pagina sin asignar, retornamos su indice
			return i
		}
	}
	return -1 
}

func AgregarPaginaACache(pagina structs.PaginaCache) {

	if !CacheHabilitado() {
		logueador.Info("Cache no habilitada, no se puede agregar pagina a cache")
		return // Cache no habilitada, no se puede agregar pagina a cache
	}
	
	if CacheLleno() {
		RemplazarPaginaEnCache(pagina) // Reemplazamos una pagina segun el algoritmo de reemplazo
		return 
	} else {
		indiceLibre := IndiceLibreCache() // Obtenemos el indice libre de la cache
		logueador.Info("Indice libre encontrado en Cache: %d", indiceLibre)
		Cache.Paginas[indiceLibre] = pagina // Asignamos la pagina al indice libre
		logueador.Info("Pagina agregada a la Cache") 
		return 
	}
}

func RemplazarPaginaEnCache(pagina structs.PaginaCache) {
	indiceVictima := IndiceDeCacheVictima() // Obtenemos el indice de la pagina victima

	if FueModificada(Cache.Paginas[indiceVictima]) { // Si la pagina victima fue modificada, debemos escribir su contenido en memoria
		logueador.Info("Pagina modificada, escribiendo en memoria")
		MandarDatosAMP(Cache.Paginas[indiceVictima]) // Enviamos la pagina a memoria
	}
	Cache.Paginas[indiceVictima] = pagina // Reemplazamos la pagina victima por la nueva pagina
	logueador.Info("Pagina reemplazada en Cache") 
}


func EscribirEnCache(pid uint, logicAdress int, data string) {

	nropagina := logicAdress / ConfigMemoria.TamanioPagina // Obtenemos el numero de pagina
	indice := ObtenerPaginaDeCache(pid, nropagina)
	if indice == -1 {
		logueador.Info("Error al obtener la pagina de Cache")
		return
	}

	offset := logicAdress % ConfigMemoria.TamanioPagina
	pagina := Cache.Paginas[indice].Contenido
	
	copy(pagina[offset:], []byte(data)) // Escribimos el contenido en la pagina de Cache
	Cache.Paginas[indice].Contenido = pagina // Actualizamos el contenido de la pagina en Cache
	Cache.Paginas[indice].BitModificado = true // Marcamos la pagina como modificada
	ActualizarRereferencia(nropagina)
	logueador.Info("Pagina escrita en Cache: PID %d, Direccion %d, Contenido %s", pid, logicAdress, data)
}

func LeerDeCache(pid uint, adress int, tam int) []byte {
	
	indice := ObtenerPaginaDeCache(pid, adress)
	
	if indice == -1 {
		logueador.Info("Println al obtener la pagina de Cache")
		return nil
	}

	if indice < 0 || indice >= len(Cache.Paginas) {
		logueador.Info("Indice de pagina fuera de rango: %d", indice)
		return nil 
	}

	pagina := Cache.Paginas[indice]
	if pagina.BitPresencia && pagina.PID == int(pid) {
		contenido := pagina.Contenido[adress:adress+tam] // Leemos el contenido de la pagina en Cache
		return contenido 
	} else {
		logueador.Info("Pagina no encontrada en Cache o no pertenece al PID %d", pid)
		return nil 
	}
}

// Para CLOCK-M
func PrimeraVueltaClockM() int {
	vueltas := 0
	for vueltas < len(Cache.Paginas) {
		i := Cache.Clock
		if !Cache.Paginas[i].BitDeUso && !Cache.Paginas[i].BitModificado {
			logueador.Info("Seleccionando pagina victima en Cache: %d - Clock en posición: %d", i, Cache.Clock)
			Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos circularmente
			return i
		}
		Cache.Clock = (Cache.Clock + 1) % len(Cache.Paginas) // Seguimos recorriendo
		vueltas++
	}
	// No se encontró ninguna página con (uso=0, modificado=0)
	return -1
}

func SegundaVueltaClockM() int {
	vueltas := 0
	for vueltas < len(Cache.Paginas) {
		i := Cache.Clock
		if !Cache.Paginas[i].BitDeUso && Cache.Paginas[i].BitModificado {
			logueador.Info("Seleccionando pagina victima en Cache: %d - Clock en posición: %d", i, Cache.Clock)
			Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos circularmente
			return i
		}
		Cache.Paginas[i].BitDeUso = false // Reiniciamos el bit de uso
		Cache.Clock = (Cache.Clock + 1) % len(Cache.Paginas) // Seguimos recorriendo
		vueltas++
	}
	return -1 // Si no se encuentra una pagina con bits 00, retorna -1
}

func BuscarVictimaClockM() int {
	for {
		valor := PrimeraVueltaClockM()
		if valor != -1 {
			return valor
		}

		valor = SegundaVueltaClockM()
		if valor != -1 {
			return valor
		}
	}
}

func Clock() int {
	for {
			i := Cache.Clock 
			if !Cache.Paginas[i].BitDeUso {
				Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
				logueador.Info("Seleccionando pagina victima en Cache: %d - Clock en posición: %d", i, Cache.Clock)
				return i
			}
			Cache.Paginas[i].BitDeUso = false  // false = 1
			Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
		}  
}

func IndiceDeCacheVictima() int {

	logueador.Info("CLOCK INICIALMENTE EN: %d", Cache.Clock)
	if Cache.Algoritmo == "CLOCK" {
		return Clock() // Llamamos a la funcion que busca la victima segun el algoritmo CLOC
	} else { // CLOCK - M
		return BuscarVictimaClockM() // Llamamos a la funcion que busca la victima segun el algoritmo CLOCK-M
	}
}

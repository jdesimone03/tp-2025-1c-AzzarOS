package utilsCPU

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"utils/logueador"
	"utils/structs"
)

var TLB structs.TLB  // Inicializamos la TLB al inicio del programa

func InicializarTLB() structs.TLB{
	if Config.TlbEntries == 0 {
		logueador.Error("TLB deshabilitado, no se inicializa")
		return structs.TLB{} // Si no hay entradas en la TLB, no se inicializa
	}

	tlb := structs.TLB{
		Entradas: make([]structs.EntradaTLB, Config.TlbEntries), // Inicializamos con un slice vacío con capacidad para TlbEntries
		MaxEntradas: Config.TlbEntries, // Establecemos el número máximo de entradas
		Algoritmo: Config.TlbReplacement, // Establecemos el algoritmo de reemplazo
	}
	for i := 0; i < Config.TlbEntries; i++ {
		tlb.Entradas = append(tlb.Entradas, structs.EntradaTLB{
			NumeroFrame: -1, // Inicialmente no hay frame asignado
			BitPresencia: false, // Inicialmente no está presente
			PID: -1, // Inicialmente no hay PID asignado
			InstanteDeReferencia: 0, // Inicialmente el instante de referencia es 0
		})
	}

	return tlb // Devolvemos la TLB inicializada
}

func AccesoATLB(pid int, nropagina int) (int, bool) {
	bool, indice := BuscarDireccion(nropagina) // Verificamos si la página está en la TLB
	if bool { 
		logueador.Info("PID: < %d > - TLB HIT - Página: %d", pid, nropagina)
		return TLB.Entradas[indice].NumeroFrame, true // Si la página está en la TLB, devolvemos el frame y true
	} else {
		logueador.Info("PID: < %d > - TLB MISS - Página: %d", pid, nropagina)
		return -1, false 
	}
}

func DesalojarTLB(pid uint) {
	logueador.Info("Desalojando TLB")
	for i := 0; i < len(TLB.Entradas); i++ {
		if TLB.Entradas[i].PID == int(pid) { // Si el PID coincide, desalojamos la entrada
		TLB.Entradas[i] = structs.EntradaTLB{
			NumeroFrame: -1, // Reinicializamos el frame
			BitPresencia: false, // Reinicializamos el bit de presencia
			PID: -1, // Reinicializamos el PID
			InstanteDeReferencia: 0, // Reinicializamos el instante de referencia
		}
	}
	}
}

// Al principio habia hecho que me devuelva la entrada victima pero era al pedo
func IndiceDeEntradaVictima() int {
		
	if TLB.Algoritmo == "FIFO" {
		return 0 
	} else { // LRU
		tiempoActual := int(time.Now().UnixNano()) // tiempo actual en nanosegundos (convertido a int)
		indice := 0
		victima := 0
		for indice < len(TLB.Entradas) {
			if TLB.Entradas[indice].InstanteDeReferencia < tiempoActual { // Si el instante de referencia es menor al tiempo actual, es una candidata a ser la victima
				victima = indice 
			}
			indice++
		}
		return victima
	}
}

func AgregarEntradaATLB(pid int, nropagina int, nroframe int) {

	nuevaEntrada := structs.EntradaTLB{
		NumeroPagina: nropagina,
		NumeroFrame: nroframe,
		BitPresencia: true, // La pagina esta presente en memoria
		PID: pid, // Asignamos el PID del proceso
		InstanteDeReferencia: int(time.Now().UnixNano()), // Asignamos el instante de referencia actual
	}

	if len(TLB.Entradas) == TLB.MaxEntradas { // si la cantidad de entradas es la maxima => hay que reemplazar
		indiceVictima := IndiceDeEntradaVictima() 
		TLB.Entradas[indiceVictima] = nuevaEntrada // reemplazo  la entrada victima por la nueva entrada
		return 
	} else { // si no esta lleno, agrego la nueva entrada al final
		TLB.Entradas = append(TLB.Entradas, nuevaEntrada)
		return 
	}
}

func BuscarDireccion(pagina int) (bool,int) { // devolvemos el frame ya que la pagina esta cargada en el TLB
	i := 0
	for i < len(TLB.Entradas) {
		if TLB.Entradas[i].NumeroPagina == pagina && TLB.Entradas[i].BitPresencia {
			return true,i // La página está en la TLB y es válida
		}
	}
	return false,-1 // La página no está en la TLB o no es válida
}

// --------------------------------------- Cache ---------------------------------------------


// Posibilidad que sea la misma estructura que la TLB
type PaginaCache struct {
	NumeroPagina int // Numero de pagina en la tabla de paginas
	BitPresencia bool // Indica si el frame esta presente en memoria
	BitModificado bool // Indica si el frame ha sido modificado
	BitDeUso bool // Indica si el frame ha sido usado recientemente
	PID int // Identificador del proceso al que pertenece el frame
	Contenido []byte // Contenido de la pagina
}

type CacheStruct struct {
	Paginas []PaginaCache 
	Algoritmo string 
	Clock int // dato para saber donde quedó la "aguja" del clock
}

var Cache CacheStruct

func InicializarCache() CacheStruct {
	return CacheStruct {
		Paginas: make([]PaginaCache, Config.CacheEntries),
		Algoritmo: Config.CacheReplacement, // FIFO o LRU => no esta en el config
	}
}

func CacheHabilitado() bool {
	return len(Cache.Paginas) > 0 
}

func FueModificada(pagina PaginaCache) bool {
	return pagina.BitModificado
}

func EstaEnCache(pid uint, nropagina int) bool {
	if !CacheHabilitado() {
		logueador.Error("Caché no habilitada, no se puede verificar si la página está en caché")
		return false 
	}

	for _, pagina := range Cache.Paginas {
		if pagina.PID == int(pid) && pagina.NumeroPagina == nropagina && pagina.BitPresencia {
			return true // La página está en la caché
		}
	}
	return false 
}

func ObtenerPaginaDeCache(pid uint, nropagina int) (int, error) {
	if !CacheHabilitado() {
		logueador.Error("Caché no habilitada, no se puede obtener la página de caché")
		return -1, fmt.Errorf("caché no habilitada")
	}

	for i, pagina := range Cache.Paginas {
		if pagina.PID == int(pid) && pagina.NumeroPagina == nropagina && pagina.BitPresencia {
			logueador.Info("Página encontrada en caché: PID %d, Página %d", pid, nropagina)
			return i, nil // Retorna la página y su índice en caché
		}
	}
	return -1, fmt.Errorf("página no encontrada en caché")
}

func MandarDatosAMP(paginas PaginaCache) {
	url := fmt.Sprintf("http://%s:%d/actualizarMP", Config.IPMemory, Config.PortMemory)
	body, err := json.Marshal(paginas)
	if err != nil {
		logueador.Info("Error al serializar la pagina a JSON:", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logueador.Error("Error al enviar la pagina a la memoria:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logueador.Error("Error al enviar la pagina a la memoria, status code: %d", resp.StatusCode)
		return
	}
	logueador.Info("Pagina enviada a la memoria correctamente")
}

func PaginasModificadas() []PaginaCache {
	var paginasModificadas []PaginaCache
	for _, pagina := range Cache.Paginas {
		if FueModificada(pagina) {
			paginasModificadas = append(paginasModificadas, pagina)
		}
	}
	return paginasModificadas
}

// Debe venir una request de memoria o kernel
func DesaolojoDeProceso(w http.ResponseWriter, r *http.Request){
	modificadas := PaginasModificadas()
	
	if len(modificadas) == 0 {
		w.Write([]byte("No hay paginas modificadas. No se actualiza memoria"))
		w.WriteHeader(http.StatusOK) // No hay paginas modificadas, todo bien
		return 
	}

	for i:=0; i < len(modificadas); i++ {
		// Consulto direccion fisica => TLB
		// contenido := modificadas[i].Contenido
		// Write de su contenido => pegarle al endpoint de memoria wirite
		// eliminar todas las entradas del caché 
		return
	}
}

func CreacionDePaginaCache(pid uint, nropagina int, contenido []byte) PaginaCache {
	return PaginaCache{
		NumeroPagina: nropagina,
		BitPresencia: true, // La pagina esta presente en memoria
		BitModificado: false, // Inicialmente no ha sido modificada
		BitDeUso: true, // Inicialmente se considera que la pagina ha sido usada
		PID: int(pid), // Asignamos el PID del proceso
		Contenido: contenido, // Asignamos el contenido de la pagina
	}
}

func PedirFrameAMemoria(pid uint, nropagina int) (PaginaCache, error) {
	
	direccionFisica := MMU(pid, nropagina) 
	url := fmt.Sprintf("http://%s:%d/pedirFrame?pid=%d&direccion=%d", Config.IPMemory, Config.PortMemory, pid, direccionFisica)
	resp, err := http.Get(url)
	if err != nil {
		logueador.Error("Error al pedir el frame a memoria: %v", err)
		return PaginaCache{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		logueador.Error("Error al pedir el frame a memoria, status code: %d", resp.StatusCode)
		return PaginaCache{}, fmt.Errorf("error al pedir el frame a memoria, status code: %d", resp.StatusCode)
	}

	var frame []byte 
	err = json.NewDecoder(resp.Body).Decode(&frame)
	if err != nil {
		logueador.Error("Error al decodificar el frame: %v", err)
		return PaginaCache{}, err
	}

	paginaCache := CreacionDePaginaCache(pid, nropagina, frame) 

	return paginaCache, nil
}

func AgregarPaginaACache(pagina PaginaCache) {
	if len(Cache.Paginas) == Config.CacheEntries {
		RemplazarPaginaEnCache(pagina) // Reemplazamos una pagina segun el algoritmo de reemplazo
		if FueModificada(pagina) {
			logueador.Info("Pagina modificada, escribiendo en memoria")
			MandarDatosAMP(pagina) 
		}
		return 
	} else {
		Cache.Paginas = append(Cache.Paginas, pagina)
		logueador.Info("Pagina agregada a la Cache") 
		return 
	}
}

func RemplazarPaginaEnCache(pagina PaginaCache) {
	indiceVictima := IndiceDeCacheVictima() // Obtenemos el indice de la pagina victima

	if FueModificada(Cache.Paginas[indiceVictima]) { // Si la pagina victima fue modificada, debemos escribir su contenido en memoria
		logueador.Info("Pagina modificada, escribiendo en memoria")
		MandarDatosAMP(Cache.Paginas[indiceVictima]) // Enviamos la pagina a memoria
	}
	Cache.Paginas[indiceVictima] = pagina // Reemplazamos la pagina victima por la nueva pagina
	logueador.Info("Pagina reemplazada en Cache") 
}


func EscribirEnCache(pid uint, adress int, data string) {

	indice, err := ObtenerPaginaDeCache(pid, adress)
	if err != nil {
		logueador.Error("Error al obtener la pagina de Cache: %v", err)
		return
	}

	Cache.Paginas[indice].Contenido = []byte(data) // Actualizamos el contenido de la pagina en Cache
	Cache.Paginas[indice].BitModificado = true // Marcamos la pagina como modificada
	logueador.Info("Pagina escrita en Cache: PID %d, Direccion %d, Contenido %s", pid, adress, data)
}

func LeerDeCache(pid uint, adress int, tam int) []byte {
	indice, err := ObtenerPaginaDeCache(pid, adress)
	if err != nil {
		logueador.Error("Error al obtener la pagina de Cache: %v", err)
		return nil
	}

	if indice < 0 || indice >= len(Cache.Paginas) {
		logueador.Error("Indice de pagina fuera de rango: %d", indice)
		return nil 
	}

	pagina := Cache.Paginas[indice]
	if pagina.BitPresencia && pagina.PID == int(pid) {
		contenido := pagina.Contenido[adress:adress+tam] // Leemos el contenido de la pagina en Cache
		return contenido 
		logueador.Info("Pagina leida de Cache: PID %d, Direccion %d, Contenido %s", pid, adress, string(contenido))
	} else {
		logueador.Error("Pagina no encontrada en Cache o no pertenece al PID %d", pid)
		return nil 
	}

	return nil 
}

// Para CLOCK-M

func IndiceDeCacheVictima() int {

	if Cache.Algoritmo == "CLOCK" {
		for {
			i := Cache.Clock 
			if !Cache.Paginas[i].BitDeUso {
				Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
				return i
			}
			Cache.Paginas[i].BitDeUso = false  // false = 1
			Cache.Clock = (i + 1) % len(Cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
		}  
	} else {
		i := 0
		for i < len(Cache.Paginas) {
			if !Cache.Paginas[i].BitDeUso && !Cache.Paginas[i].BitModificado {
				Cache.Paginas[i].BitDeUso = true 
				return i // Retorna el indice de la primera pagina con bits 00
			} else {
				if !Cache.Paginas[i].BitDeUso && Cache.Paginas[i].BitModificado { 
					Cache.Paginas[i].BitDeUso = true
					return i;
				}
			}
		}
	}
	return -1 // Si no se encuentra una pagina con bits 00, retorna -1
}


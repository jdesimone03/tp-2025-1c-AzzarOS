package utilsCPU

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"utils/logueador"
	"utils/structs"
)
// --------------------------------------- Cache ---------------------------------------------


var Cache structs.CacheStruct = InicializarCache()

func InicializarCache() structs.CacheStruct {
	paginas := make([]structs.PaginaCache, Config.CacheEntries) // Slice vacío, capacidad predefinida
	
		for i := 0; i < Config.CacheEntries; i++ {
		paginas[i] = structs.PaginaCache{
			NumeroPagina: -1,
			PID: -1,
			}
		}
	return structs.CacheStruct{
		Paginas: paginas,
		Algoritmo: Config.CacheReplacement,
	}
}

func CacheHabilitado() bool {
	return len(Cache.Paginas) > 0 
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

func PaginasModificadas() []structs.PaginaCache {
	var paginasModificadas []structs.PaginaCache
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

func EliminarEntradasDeCache(pid uint) {
	logueador.Info("Eliminando entradas de caché para el PID %d", pid)
	for i := 0; i < len(Cache.Paginas); i++ {
		if Cache.Paginas[i].PID == int(pid) { // Si el PID coincide, eliminamos la entrada
			Cache.Paginas[i] = structs.PaginaCache{} // Reinicializamos la entrada
			logueador.Info("Entrada de caché eliminada para el PID %d", pid)
		}
	}
}

func CreacionDePaginaCache(pid uint, nropagina int, contenido []byte) structs.PaginaCache {
	return structs.PaginaCache{
		NumeroPagina: nropagina,
		BitPresencia: true, // La pagina esta presente en memoria
		BitModificado: false, // Inicialmente no ha sido modificada
		BitDeUso: true, // Inicialmente se considera que la pagina ha sido usada
		PID: int(pid), // Asignamos el PID del proceso
		Contenido: contenido, // Asignamos el contenido de la pagina
	}
}

func PedirFrameAMemoria(pid uint, nropagina int) (structs.PaginaCache, error) {
	
	direccionFisica := MMU(pid, nropagina) 
	url := fmt.Sprintf("http://%s:%d/pedirFrame?pid=%d&direccion=%d", Config.IPMemory, Config.PortMemory, pid, direccionFisica)
	
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

	paginaCache := CreacionDePaginaCache(pid, nropagina, frame) 

	return paginaCache, nil
}

func AgregarPaginaACache(pagina structs.PaginaCache) {
	
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


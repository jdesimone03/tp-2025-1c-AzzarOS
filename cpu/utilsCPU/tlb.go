package utilsCPU

import (
	"log"
	"time"
	"fmt"
	"encoding/json"
	"bytes"
	"net/http"
)

type EntradaTLB struct {
	NumeroPagina int
	NumeroFrame int 
	BitPresencia bool // Indica si el frame esta presente en memoria
	BitModificado bool // Indica si el frame ha sido modificado
	PID int // Identificador del proceso al que pertenece el frame
	InstanteDeReferencia int // Marca el instante de referencia para LRU
}

// Algoritmos => FIFO o LRU
// TLB[Pagina] => marco
// Primera opción:  

var tlb1 = make([]EntradaTLB, Config.TlbEntries)

func InicializarTLB() {
	if Config.TlbEntries == 0 {
		log.Println("TLB deshabilitado, no se inicializa")
		return
	}
		for i := 0; i < Config.TlbEntries; i++ {
		tlb1[i] = EntradaTLB{
			NumeroFrame: -1, // Inicialmente no hay frame asignado
			BitPresencia: false, // Inicialmente no está presente
			BitModificado: false, // Inicialmente no ha sido modificado
			PID: -1, // Inicialmente no hay PID asignado
			InstanteDeReferencia: 0, // Inicialmente el instante de referencia es 0
		}
	}
}

func TLBLleno() bool {
	for _, entrada := range tlb1 {
		if entrada.NumeroFrame == -1 { // Si hay al menos una entrada sin asignar
			return false
		}
	}
	return true // Todas las entradas están ocupadas
}

// Segunda opción:
type TLB struct {
	Entradas []EntradaTLB 	
	MaxEntradas int
	Algoritmo string 
}

var tlb = TLB {
	Entradas: make([]EntradaTLB, Config.TlbEntries),
	MaxEntradas: Config.TlbEntries,
	Algoritmo: Config.TlbReplacement, // FIFO o LRU => no esta en el config
}

func InicializarTLBStruct() {
	if Config.TlbEntries == 0 {
		log.Println("TLB deshabilitado, no se inicializa")
		return
	}
	for i := 0; i < Config.TlbEntries; i++ {
		tlb.Entradas = append(tlb.Entradas, EntradaTLB{
			NumeroFrame: -1, // Inicialmente no hay frame asignado
			BitPresencia: false, // Inicialmente no está presente
			BitModificado: false, // Inicialmente no ha sido modificado
			PID: -1, // Inicialmente no hay PID asignado
			InstanteDeReferencia: 0, // Inicialmente el instante de referencia es 0
		})
	}
}

func AccesoATLB(pid int, nropagina int) (int, bool) {
	bool, indice := Hit(nropagina) // Verificamos si la página está en la TLB
	if bool { 
		log.Println("PID: <", pid, "> - TLB HIT - Página:", nropagina)
		return tlb.Entradas[indice].NumeroFrame, true // Si la página está en la TLB, devolvemos el frame y true
	} else {
		log.Println("PID: <", pid, "> - TLB MISS - Página:", nropagina)
		return
		// Se busca en la tabla de paginas => GET a memoria 
		// Se remplaza
		// Se devuelve el frame 
	}
}


// Caso hit => la página está en la TLB
// Caso miss => la página no está en la TLB, se debe buscar en la memoria principal y agregar a la TLB

// FIFO => victima => la que más tiempo lleva en el TLB
// LRU => victima => la que menos tiempo ha sido referenciada

// Al principio habia hecho que me devuelva la entrada victima pero era al pedo
func IndiceDeEntradaVictima() int {
	var victima EntradaTLB
	entradas := tlb.Entradas

	if tlb.Algoritmo == "FIFO" {
		victima = entradas[0] // Seguro la primera vez es la mas vieja, pero a medida que vamos sacando entradas, puede que no sea la más vieja
		indice := 1 // si lo pongo como variable en el for, no puedo devoverlo 
		for indice < len(entradas) {
			if entradas[indice].InstanteDeReferencia < victima.InstanteDeReferencia { // El que entro en el instante 0 tiene que ser el más viejo
				victima = entradas[indice]
			}
			indice++
		}
		return indice
	} else { // no dice mas algoritmos => LRU 
		tiempoActual := int(time.Now().UnixNano()) // tiempo actual en nanosegundos (convertido a int)
		indice := 0
		for indice < len(entradas) {
			if entradas[indice].InstanteDeReferencia < tiempoActual { // Si el instante de referencia es menor al tiempo actual, es una candidata a ser la victima
				victima = entradas[indice]
			}
			indice++
		}
		return indice
	}

}

func ReemplzarEntradaTLB(pid int, nropagina int, nroframe int) {
	
	if len(tlb.Entradas) == 0 {
		log.Println("No hay entradas en la TLB para reemplazar")
		return
	}

	nuevaEntrada := EntradaTLB{
		NumeroPagina: nropagina,
		NumeroFrame: nroframe,
		BitPresencia: true, // La pagina esta presente en la TLB	
		BitModificado: false, // Inicialmente no ha sido modificada
		PID: pid, // Asignamos el PID del proceso
		InstanteDeReferencia: int(time.Now().UnixNano()), // Asignamos el instante de referencia actual
	}

	if len(tlb.Entradas) == tlb.MaxEntradas { // si la cantidad de entradas es la maxima => hay que reemplazar
		indiceVictima := IndiceDeEntradaVictima() 
		tlb.Entradas[indiceVictima] = nuevaEntrada // reemplazo  la entrada victima por la nueva entrada
		return 
	} else { // si no esta lleno, agrego la nueva entrada al final
		tlb.Entradas = append(tlb.Entradas, nuevaEntrada)
		return 
	}
}

func Hit(pagina int) (bool,int) { // devolvemos el frame ya que la pagina esta cargada en el TLB
	i := 0
	for i < len(tlb.Entradas) {
		if tlb.Entradas[i].NumeroPagina == pagina && tlb.Entradas[i].BitPresencia {
			return true,i // La página está en la TLB y es válida
		}
	}
	return false,-1 // La página no está en la TLB o no es válida
}

// --------------------------------------- Cache ---------------------------------------------

/*
A la hora de modificar una página:
	1. Se debe corroborar si la Cache esta habilitada => tiene al menos 1 frame
		De ser asi: 
			a. Se hacen las operaciones en caché 
		Si no esat habilitada:
			b. Se hace un write a memoria directamente


A la hora de cargar una página en cache
	1. Hay que corroborar si se encuentra llena 
		De ser asi:
			a. Se debe reemplazar una página según el algoritmo de reemplazo
			b. Si la página fue modificada, los cambios deben ser escritos en memoria


Al la hora de desojar un proceso: 
	1. Las páginas que se encuentran modificadas deben ser actualizadas en memoria principal
		a. Primero se consultan las direcciones fisicas 
		b. Se envian a escribir su contenido a memoria 
		c. Se eliminan todas las entradas de la caché
	
	
Para acceder a una página hay que:
1. Verificar que este en caché
2. Despues se pasa a la TLB
3. Como ultima instancia la tabla de paginas en memoria 

Preguntas: 
1) Se cargan todas las paginas en cache de un proceso al iniciar?


*/

// Posibilidad que sea la misma estructura que la TLB
type PaginaCache struct {
	NumeroFrame int // Numero de pagina
	BitPresencia bool // Indica si el frame esta presente en memoria
	BitModificado bool // Indica si el frame ha sido modificado
	BitDeUso bool // Indica si el frame ha sido usado recientemente
	PID int // Identificador del proceso al que pertenece el frame
	Contenido []byte // Contenido de la pagina
}


type Cache struct {
	Paginas []PaginaCache
	Algoritmo string 
	Clock int // dato para saber donde quedó la "aguja" del clock
}

var cache = InicializarCache()

func InicializarCache() Cache {
	return Cache {
		Paginas: make([]PaginaCache, Config.CacheEntries),
		Algoritmo: Config.CacheReplacement, // FIFO o LRU => no esta en el config
	}
}

func CacheHabilitado() bool {
	return len(cache.Paginas) > 0 
}

func FueModificada(pagina PaginaCache) bool {
	return pagina.BitModificado
}

func EstaEnCache(pid uint, nropagina int) bool {
	if !CacheHabilitado() {
		return false 
	}

	for _, pagina := range cache.Paginas {
		if pagina.PID == int(pid) {
			return true // La página está en la caché
		}
	}
	return false 
}

// mandamos una lista de paginas cache para reutilizar en el codigo 
func MandarDatosAMP(paginas []PaginaCache) {
	url := fmt.Sprintf("http://%s:%d/cache", Config.IPMemory, Config.PortMemory)
	body, err := json.Marshal(paginas)
	if err != nil {
		log.Println("Error al serializar la pagina a JSON:", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Println("Error al enviar la pagina a la memoria:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error al enviar la pagina a la memoria, status code: %d", resp.StatusCode)
		return
	}
	log.Println("Pagina enviada a la memoria correctamente")
}

func PaginasModificadas() []PaginaCache {
	var paginasModificadas []PaginaCache
	for _, pagina := range cache.Paginas {
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

func AgregarPaginaACache(pagina PaginaCache) {
	
	if len(cache.Paginas) == Config.CacheEntries {
		// Remplazar segun el algoritmo de reemplazo
	} else {
		cache.Paginas = append(cache.Paginas, pagina)
		log.Println("Pagina agregada a la cache")
	}
}

// Para CLOCK-M

func IndiceDeCacheVictima() int {

	if cache.Algoritmo == "CLOCK" {
		for {
			i := cache.Clock 
			if !cache.Paginas[i].BitDeUso {
				cache.Clock = (i + 1) % len(cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
				return i
			}
			cache.Paginas[i].BitDeUso = false  // false = 1
			cache.Clock = (i + 1) % len(cache.Paginas) // Avanzamos al siguiente indice circularmente => por si llegamos al final del vector, poder volver al inicio
		}  
	} else {
		i := 0
		for i < len(cache.Paginas) {
			if !cache.Paginas[i].BitDeUso && !cache.Paginas[i].BitModificado {
				cache.Paginas[i].BitDeUso = true 
				return i // Retorna el indice de la primera pagina con bits 00
			} else {
				if !cache.Paginas[i].BitDeUso && cache.Paginas[i].BitModificado { 
					cache.Paginas[i].BitDeUso = true
					return i;
				}
			}
		}
	}
}

func AccederACache(numeroPagina int) (PaginaCache, bool) {
	if !CacheHabilitado(){
		return PaginaCache{}, false // La cache no está habilitada
		// mandar mensaje a memoria para que haga las operaciones
	} else {
			return 
	}
}
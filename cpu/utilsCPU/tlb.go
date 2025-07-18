package utilsCPU

import (
	"time"
	"utils/logueador"
	"utils/structs"

	// "golang.org/x/tools/go/analysis/passes/appends"
)

var tlb structs.TLB 

func DesalojarTLB(pid uint) {
	
	logueador.Info("Desalojando TLB")
	for i := 0; i < len(tlb.Entradas); i++ {
		if tlb.Entradas[i].PID == int(pid) { // Si el PID coincide, desalojamos la entrada
		tlb.Entradas[i] = structs.EntradaTLB{
			NumeroFrame: -1, // Reinicializamos el frame
			NumeroPagina: -1, // Reinicializamos el número de página
			BitPresencia: false, // Reinicializamos el bit de presencia
			PID: -1, // Reinicializamos el PID
			Llegada: -1, // Reinicializamos el instante de referencia
			Referencia: -1, // Reinicializamos la referencia
		}
	}
	}
}

func InicializarTLB() {
	entradas := make([]structs.EntradaTLB, Config.TlbEntries)
	
	for i := 0; i < Config.TlbEntries; i++ {
		entradas[i] = structs.EntradaTLB{
			NumeroPagina:        -1, // Inicialmente no hay páginas cargadas
			NumeroFrame:         -1, // Inicialmente no hay frames asignados
			BitPresencia:        false, // Inicialmente no hay páginas presentes
			PID:                 -1, // Inicialmente no hay PID asignado
			Llegada: -1, // Inicialmente no hay instante de referencia
			Referencia: -1, // Inicialmente no hay referencia
		}
	}
	tlb = structs.TLB{
		Entradas:    entradas,
		MaxEntradas: Config.CacheEntries,
		Algoritmo:   Config.CacheReplacement,
	}
}


func TLBHabilitada() bool {
	return Config.TlbEntries > 0
}

func BuscarDireccion(pagina int) (bool,int) { // devolvemos el frame ya que la pagina esta cargada en el TLB
	
	for i := 0; i < len(tlb.Entradas);i++ {
		if tlb.Entradas[i].NumeroPagina == pagina && tlb.Entradas[i].BitPresencia {
			return true,i // La página está en la TLB y es válida
		}
	}
	return false,-1 // La página no está en la TLB o no es válida
}

func ActualizarReferencia(indice int) {
	entradaReferenciada := tlb.Entradas[indice]
	entradaReferenciada.Referencia = int(time.Now().UnixNano()) // Actualizamos la referencia al instante actual
	tlb.Entradas[indice] = entradaReferenciada // Actualizamos la entrada en la TLB
}

func AccesoATLB(pid int, nropagina int) int {
	
	if !TLBHabilitada() {
		logueador.Warn("TLB no habilitada, no se puede acceder a la TLB")
		return -1 // TLB no habilitada, no se puede acceder a la
	}
	
	bool, indice := BuscarDireccion(nropagina) // Verificamos si la página está en la TLB
	if bool { 
		logueador.TLBHit(uint(pid), nropagina)
		ActualizarReferencia(indice) // Actualizamos la referencia al instante actual
		MostrarContenidoTLB()
		return tlb.Entradas[indice].NumeroFrame // Si la página está en la TLB, devolvemos el frame y true
	} else {
		logueador.TLBMiss(uint(pid), nropagina)
		MostrarContenidoTLB()
		return -1
	}
}


// "segun" es una función que recibe una entrada de TLB y devuelve un entero que representa el criterio de selección para la víctima.
func IndiceDeEntradaVictima(segun func(structs.EntradaTLB) int) int { // Según llegada o referencia, dependiendo del algoritmo de reemplazo
		
	victima := tlb.Entradas[0] // Inicializamos la víctima con la primera entrada
	indice := 0
	for i:=0; i < len(tlb.Entradas); i++{
		if segun(victima) > segun(tlb.Entradas[i]) { // Comparamos la entrada actual con la víctima
			victima = tlb.Entradas[i] // Si la entrada actual es más
			indice = i // Actualizamos el índice de la víctima
		}

	}
	return indice 
}

func TLBLleno() bool {
	return len(tlb.Entradas) == Config.TlbEntries // Verificamos si la TLB está llena
}

func EntradaTLBValida() int {
	for i := 0; i < len(tlb.Entradas); i++ {
		if tlb.Entradas[i].PID == -1 { // Si hay una entrada con número de página -1, significa que la TLB tiene espacio
			return i // Retorna el índice de la entrada válida
		}
	}
	return -1 // Si no hay entradas válidas, retorna -1
}


func AgregarEntradaATLB(pid int, nropagina int, nroframe int) {

	if !TLBHabilitada() {
		logueador.Warn("TLB no habilitada, no se puede agregar una entrada a la TLB")
		return // TLB no habilitada, no se puede agregar una entrada
	}

	nuevaEntrada := structs.EntradaTLB{
		NumeroPagina: nropagina,
		NumeroFrame: nroframe,
		BitPresencia: true, // La pagina esta presente en memoria
		PID: pid, // Asignamos el PID del proceso
		Llegada: int(time.Now().UnixNano()), // Asignamos el instante de referencia actual
		Referencia: int(time.Now().UnixNano()), // Asignamos la referencia actual
	}

	if TLBLleno() { 
		indiceVictima := IndiceDeEntradaVictima(func(e structs.EntradaTLB) int {
			if Config.TlbReplacement == "FIFO" {
				return e.Llegada
			} else { // Como solamente dos algoritmos, el otro tiene que ser LRU 
				return e.Referencia 
			}
		}) 
		tlb.Entradas[indiceVictima] = nuevaEntrada // reemplazo  la entrada victima por la nueva entrada
		return 
	} else { // si no esta lleno, agrego la nueva entrada al final
		indiceValido := EntradaTLBValida() // Buscamos un indice valido para agregar la nueva entrada
		logueador.Debug("Indice libre en TLB: %d", indiceValido)
		tlb.Entradas[indiceValido] = nuevaEntrada // Asignamos la nueva entrada al indice valido
		logueador.Info("Agregando entrada a TLB - PID: %d, Página: %d, Frame: %d", pid, nropagina, nroframe)
		return 
	}
}
func DesalojoTlB(pid uint) {
	for i := 0; i < len(tlb.Entradas); i++ {
		if tlb.Entradas[i].PID == int(pid) { // Verificamos si la entrada pertenece al PID
			tlb.Entradas[i] = structs.EntradaTLB{ // Desalojamos la entrada del PID
				NumeroPagina: -1, // -1 indica que la entrada está vacía
				NumeroFrame:  -1,
				BitPresencia: false,
				PID:          -1,
				Referencia: -1,
				Llegada: -1, // -1 indica que la entrada no ha sido utilizada
			}
		}
	}
}

func MostrarContenidoTLB() {
	logueador.Info("-------------------- TLB CONTENIDO ------------------------------")
	for i, entrada := range tlb.Entradas {
		logueador.Info("Entrada %d: PID=%d, Pagina=%d, Frame=%d, Presente=%v, Llegada=%d, Referencia=%d",
			i, entrada.PID, entrada.NumeroPagina, entrada.NumeroFrame, entrada.BitPresencia, entrada.Llegada, entrada.Referencia)
	}
	logueador.Info("------------------------------------------------------------------")
}
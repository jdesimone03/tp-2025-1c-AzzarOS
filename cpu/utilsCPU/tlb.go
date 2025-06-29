package utilsCPU

import (
	"utils/logueador"
	"time"
	"utils/structs"
)

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

func InicializarTLB() structs.TLB {
	entradas := make([]structs.EntradaTLB, 0, Config.CacheEntries)
	for i := 0; i < Config.CacheEntries; i++ {
		entradas = append(entradas, structs.EntradaTLB{
			NumeroPagina:        -1, // Inicialmente no hay páginas cargadas
			NumeroFrame:         -1, // Inicialmente no hay frames asignados
			BitPresencia:        false, // Inicialmente no hay páginas presentes
			PID:                 -1, // Inicialmente no hay PID asignado
			Llegada: -1, // Inicialmente no hay instante de referencia
			Referencia: -1, // Inicialmente no hay referencia
		})
	}
	return structs.TLB{
		Entradas:    entradas,
		MaxEntradas: Config.CacheEntries,
		Algoritmo:   Config.CacheReplacement,
	}
}

var tlb structs.TLB = InicializarTLB()

func TLBHabilitada() bool {
	return tlb.MaxEntradas != 0
}

func BuscarDireccion(pagina int) (bool,int) { // devolvemos el frame ya que la pagina esta cargada en el TLB
	
	for i := 0; i < len(tlb.Entradas);i++ {
		if tlb.Entradas[i].NumeroPagina == pagina && tlb.Entradas[i].BitPresencia {
			return true,i // La página está en la TLB y es válida
		}
	}
	return false,-1 // La página no está en la TLB o no es válida
}

func ActualizarRereferencia(indice int ) {
	entradaReferenciada := tlb.Entradas[indice]
	entradaReferenciada.Referencia = int(time.Now().UnixNano()) // Actualizamos la referencia al instante actual
	tlb.Entradas[indice] = entradaReferenciada // Actualizamos la entrada en la TLB
	return 
}

func AccesoATLB(pid int, nropagina int) int {
	
	if !TLBHabilitada() {
		logueador.Info("TLB no habilitada, no se puede acceder a la TLB")
		return -1 // TLB no habilitada, no se puede acceder a la
	}
	
	bool, indice := BuscarDireccion(nropagina) // Verificamos si la página está en la TLB
	if bool { 
		logueador.Info("PID: < %d > - TLB HIT - Página: %d", pid, nropagina)
		ActualizarRereferencia(indice) // Actualizamos la referencia al instante actual
		return tlb.Entradas[indice].NumeroFrame // Si la página está en la TLB, devolvemos el frame y true
	} else {
		logueador.Info("PID: < %d > - TLB MISS - Página: %d", pid, nropagina)
		return -1
	}
}

func IndiceDeEntradaVictima(segun func(structs.EntradaTLB) int) int {
		
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
	for i:= 0; i < len(tlb.Entradas); i++ {
		if tlb.Entradas[i].NumeroPagina == -1 { // Si hay una entrada con número de página -1, significa que la TLB no está llena
			return false // La TLB no está llena
		}
	}
	return true 
}

func EntradaTLBValida() int {
	for i := 0; i < len(tlb.Entradas); i++ {
		if tlb.Entradas[i].NumeroPagina == -1 { // Si hay una entrada con número de página -1, significa que la TLB tiene espacio
			return i // Retorna el índice de la entrada válida
		}
	}
	return -1 // Si no hay entradas válidas, retorna -1
}


func AgregarEntradaATLB(pid int, nropagina int, nroframe int) {

	nuevaEntrada := structs.EntradaTLB{
		NumeroPagina: nropagina,
		NumeroFrame: nroframe,
		BitPresencia: true, // La pagina esta presente en memoria
		PID: pid, // Asignamos el PID del proceso
		Llegada: int(time.Now().UnixNano()), // Asignamos el instante de referencia actual
	}

	if TLBLleno() { 
		indiceVictima := IndiceDeEntradaVictima(func(e structs.EntradaTLB) int {
			if Config.CacheReplacement == "FIFO" {
				return e.Llegada
			} else {
				return e.Referencia 
			}
		}) 
		tlb.Entradas[indiceVictima] = nuevaEntrada // reemplazo  la entrada victima por la nueva entrada
		return 
	} else { // si no esta lleno, agrego la nueva entrada al final
		indiceValido := EntradaTLBValida() // Buscamos un indice valido para agregar la nueva entrada
		tlb.Entradas[indiceValido] = nuevaEntrada // Asignamos la nueva entrada al indice valido
		logueador.Info("Agregando entrada a TLB - PID: %d, Página: %d, Frame: %d", pid, nropagina, nroframe)
		return 
	}
}

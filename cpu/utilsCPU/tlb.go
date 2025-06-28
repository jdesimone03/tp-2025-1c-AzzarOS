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
			BitPresencia: false, // Reinicializamos el bit de presencia
			PID: -1, // Reinicializamos el PID
			InstanteDeReferencia: 0, // Reinicializamos el instante de referencia
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
			InstanteDeReferencia: 0, // Inicialmente no hay instante de referencia
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


func AccesoATLB(pid int, nropagina int) int {
	
	if !TLBHabilitada() {
		logueador.Info("TLB no habilitada, no se puede acceder a la TLB")
		return -1 // TLB no habilitada, no se puede acceder a la
	}
	
	bool, indice := BuscarDireccion(nropagina) // Verificamos si la página está en la TLB
	if bool { 
		logueador.Info("PID: < %d > - TLB HIT - Página: %d", pid, nropagina)
		return tlb.Entradas[indice].NumeroFrame // Si la página está en la TLB, devolvemos el frame y true
	} else {
		logueador.Info("PID: < %d > - TLB MISS - Página: %d", pid, nropagina)
		return -1
	}
}

func IndiceDeEntradaVictima() int {
		
	if tlb.Algoritmo == "FIFO" {
		return 0 
	} else { // LRU
		tiempoActual := int(time.Now().UnixNano()) // tiempo actual en nanosegundos (convertido a int)
		indice := 0
		victima := 0
		for indice < len(tlb.Entradas) {
			if tlb.Entradas[indice].InstanteDeReferencia < tiempoActual { // Si el instante de referencia es menor al tiempo actual, es una candidata a ser la victima
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

	if len(tlb.Entradas) == tlb.MaxEntradas { // si la cantidad de entradas es la maxima => hay que reemplazar
		indiceVictima := IndiceDeEntradaVictima() 
		tlb.Entradas[indiceVictima] = nuevaEntrada // reemplazo  la entrada victima por la nueva entrada
		return 
	} else { // si no esta lleno, agrego la nueva entrada al final
		tlb.Entradas = append(tlb.Entradas, nuevaEntrada)
		return 
	}
}

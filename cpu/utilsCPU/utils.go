package utilsCPU

import (
	"strconv"
	"utils/logueador"
	"utils/structs"
	"fmt"
	"net/http"
	"encoding/json"
	"math"
)

// -------------------------------- MMU --------------------------------- //

func PedirConfigMemoria() error  {
	url := fmt.Sprintf("http://%s:%s/config", Config.IPMemory, Config.PortMemory)
	logueador.Info("Solicitando configuración de Memoria en: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("falló el GET a memoria: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("memoria respondió con error HTTP %d", resp.StatusCode)
	}

	var config structs.ConfigMemoria
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("falló el decode del JSON: %w", err)
	}

	ConfigMemoria = &config
	return nil
}

var ConfigMemoria *structs.ConfigMemoria


func nroPagina(direccionLogica int, pagesize int) int {
	return direccionLogica / pagesize
}

func desplazamiento(direccionLogica int, pagesize int) int {
	return direccionLogica % pagesize
}

func entradaNiveln(direccionlogica int, niveles int, idTabla int) int {
	pagina := direccionlogica / ConfigMemoria.TamanioPagina
	divisor := int(math.Pow(float64(ConfigMemoria.EntradasPorTabla), float64(niveles - idTabla)))
	return (pagina / divisor) % ConfigMemoria.EntradasPorTabla
}

func PedirTablaDePaginas(pid uint) *structs.Tabla {
	url := fmt.Sprintf("http://%s:%s/tabla-paginas?pid=", Config.IPMemory, Config.PortMemory) + strconv.Itoa(int(pid))
	resp, err := http.Get(url)
	if err != nil {
		logueador.Error("Error al solicitar la tabla de páginas: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logueador.Error("Error al obtener la tabla de páginas, código de estado: %d", resp.StatusCode)
		return nil
	}

	var tabla structs.Tabla
	if err := json.NewDecoder(resp.Body).Decode(&tabla); err != nil {
		logueador.Error("Error al decodificar la tabla de páginas: %v", err)
		return nil
	}

	return &tabla
}

func MMU(pid uint, direccionLogica int) int {

	desplazamiento := desplazamiento(direccionLogica, ConfigMemoria.TamanioPagina)
	tabla := PedirTablaDePaginas(pid) // Obtengo la tabla de páginas del PID

	if tabla == nil {	
		logueador.Info("No se pudo obtener la tabla de páginas para el PID %d", pid)
		return -1
	}
	
	raiz := tabla
	for nivel := 1; nivel <= ConfigMemoria.CantNiveles; nivel++ {
		entrada := entradaNiveln(direccionLogica, ConfigMemoria.CantNiveles, nivel)
		// Si llegamos al nivel final => queda buscar el frame unicamente 
		if nivel == ConfigMemoria.CantNiveles {
			if entrada >= len(raiz.Valores) || raiz.Valores[entrada] == -1 { // verifico si la entrada es válida
				logueador.Info("Dirección lógica %d no está mapeada en la tabla de páginas del PID %d", direccionLogica, pid)
				return -1 // Dirección no mapeada
			}
		frame := raiz.Valores[entrada] // Obtengo el frame correspondiente a la entrada
		return frame*ConfigMemoria.TamanioPagina + desplazamiento // Esto es el frame correspondiente a la dirección lógica 
		}
		// Si estamos en niveles intermedios => seguimos recorriendo la tabla de páginas
		if entrada >= len(raiz.Punteros) || raiz.Punteros[entrada] == nil { 
			logueador.Info("Dirección lógica %d no está mapeada en la tabla de páginas del PID %d", direccionLogica, pid)
			return -1 // Dirección no mapeada
		}
		raiz = raiz.Punteros[entrada] // Avanzamos al siguiente nivel de la tabla de páginas
	}

	logueador.Info("Error al procesar la dirección lógica %d para el PID %d", direccionLogica, pid)
	return -1 // Si llegamos hasta aca => error en el procesamiento de la dirección lógica
}

// ---------------------------------- TRADUCCIÓN DE DIRECCIONES ----------------------------------//

func TraducirDireccion(pid uint, direccion int) int {

	logueador.Info("Traduciendo dirección lógica a física")
	paginaLogica := direccion / ConfigMemoria.TamanioPagina 
	offset := desplazamiento(direccion, ConfigMemoria.TamanioPagina) // desplazamiento dentro de la página

	logueador.Info("Accediendo a TLB")
	// 1. Preguntamos a TLB
	frame := AccesoATLB(int(pid), paginaLogica) // Verificamos si la página está en la TLB
	if frame != -1 {
		return frame * ConfigMemoria.TamanioPagina + offset // Retornamos la dirección física
	} 
	logueador.Info("Página no encontrada en TLB, buscando en tabla de páginas - MMU")
	// 2. Si no está en TLB, buscamos en la tabla de páginas
	direccionFisica := MMU(pid, direccion) // Obtenemos el frame físico correspondiente a la página lógica
	if direccionFisica == -1 {
		logueador.Info("Error al traducir la dirección lógica %d para el PID %d", direccion, pid)
		return -1 // Retornamos -1 para indicar que no se pudo traducir la dirección
	}

	frameFisico := direccionFisica / ConfigMemoria.TamanioPagina 

	// HUBO MISS => AGREGAR A TLB
	AgregarEntradaATLB(int(pid), paginaLogica, frameFisico) // Agregamos la entrada a la TLB
	return direccionFisica// Retornamos la dirección física
}

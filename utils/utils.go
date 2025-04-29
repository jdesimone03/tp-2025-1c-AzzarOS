package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"utils/structs"	
	"strconv"
	"strings"
)

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string) {
	logFile, err := os.OpenFile(nombreArchivoLog+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // FALTO CAMBIAR A SLOG, TE LO DEJO A VOS FABRI
	slog.Info("Logger " + nombreArchivoLog + ".log configurado")
}

func EnviarMensaje(ip string, puerto int, endpoint string, mensaje any) {
	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo codificar el mensaje (%v)", err))
		return
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo enviar mensaje a %s:%d/%s (%v)", ip, puerto, endpoint, err))
		return
	}

	// log.Printf("respuesta del servidor: %s", resp.Status)
	slog.Info(fmt.Sprintf("Respuesta de %s:%d/%s %v", ip, puerto, endpoint, resp.Status))
}

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[string](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Info(fmt.Sprintf("Me llego un mensaje: %+v",mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirPCB(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[structs.PCB](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Info(fmt.Sprintf("Me llego un mensaje: %+v",mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Función genérica para decodificar un mensaje del body
func DecodificarMensaje[T any](r *http.Request) (*T, error) {
	var mensaje T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&mensaje)
	if err != nil {
		return nil, err
	}
	return &mensaje, nil
}

func IniciarServidor(puerto int) error {
	err := http.ListenAndServe(fmt.Sprintf(":%d",puerto), nil)
	if err != nil {
		panic(err)
	}
	return err
}


// log obligatorio para cuando se crea un nuevo proceso (1)
func NuevoProceso(nuevoPCB structs.PCB) {
	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}

// PROPUESTA FUNCION PARSEO DE COMANDOS

// Mapa para pasar de  string a InstructionType (nos serviria para el parsing)
var instructionMap = map[string]structs.InstructionType{
    "NOOP":        structs.INST_NOOP,
    "WRITE":       structs.INST_WRITE,
    "READ":        structs.INST_READ,
    "GOTO":        structs.INST_GOTO,
    "IO":          structs.INST_IO,
    "INIT_PROC":   structs.INST_INIT_PROC,
    "DUMP_MEMORY": structs.INST_DUMP_MEMORY,
    "EXIT":        structs.INST_EXIT,
}

//Tengo que cambiar los logs
func parseLine(line string) (interface{}, error) {
    parts := strings.Fields(line) // Divide por espacios
    if len(parts) == 0 {
        return nil, fmt.Errorf("línea vacía")
    }

    cmd := parts[0]
    params := parts[1:]

    instType, ok := instructionMap[cmd]
    if !ok {
        return nil, fmt.Errorf("comando desconocido: %s", cmd)
    }

    switch instType {
    case structs.INST_NOOP:
        if len(params) != 0 { return nil, fmt.Errorf("NOOP no espera parámetros") }
        return structs.NoopInstruction{}, nil

    case structs.INST_WRITE:
        if len(params) != 2 { return nil, fmt.Errorf("WRITE espera 2 parámetros (Dirección, Datos)") }
        addr, err := strconv.Atoi(params[0])
        if err != nil { return nil, fmt.Errorf("parámetro Dirección inválido para WRITE: %v", err) }
        return structs.WriteInstruction{Address: addr, Data: params[1]}, nil

     case structs.INST_READ:
         if len(params) != 2 { return nil, fmt.Errorf("READ espera 2 parámetros (Dirección, Tamaño)") }
         addr, err := strconv.Atoi(params[0])
         if err != nil { return nil, fmt.Errorf("parámetro Dirección inválido para READ: %v", err) }
         size, err := strconv.Atoi(params[1])
         if err != nil { return nil, fmt.Errorf("parámetro Tamaño inválido para READ: %v", err) }
         return structs.ReadInstruction{Address: addr, Size: size}, nil

     case structs.INST_GOTO:
         if len(params) != 1 { return nil, fmt.Errorf("GOTO espera 1 parámetro (Valor)") }
         target, err := strconv.Atoi(params[0])
         if err != nil { return nil, fmt.Errorf("parámetro Valor inválido para GOTO: %v", err) }
         return structs.GotoInstruction{TargetAddress: target}, nil

     case structs.INST_DUMP_MEMORY:
         if len(params) != 0 { return nil, fmt.Errorf("DUMP_MEMORY no espera parámetros") }
         return structs.DumpMemoryInstruction{}, nil

     case structs.INST_EXIT:
        if len(params) != 0 { return nil, fmt.Errorf("EXIT no espera parámetros") }
        return structs.ExitInstruction{}, nil
    default:
        return nil, fmt.Errorf("parsing no implementado para: %s", cmd)
    }
}
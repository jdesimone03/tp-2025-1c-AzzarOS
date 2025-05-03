package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"utils/structs"
)

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string) {
	logFile, err := os.OpenFile(nombreArchivoLog+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	slog.Info("Logger " + nombreArchivoLog + ".log configurado")
}

func EnviarMensaje(ip string, puerto int, endpoint string, mensaje any) string {
	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo codificar el mensaje (%v)", err))
		return ""
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo enviar mensaje a %s:%d/%s (%v)", ip, puerto, endpoint, err))
		return ""
	}
	defer resp.Body.Close()

	var resData structs.Respuesta
	respuesta, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	err = json.Unmarshal(respuesta, &resData)
	if err != nil {
		return ""
	}

	// log.Printf("respuesta del servidor: %s", resp.Status)
	slog.Info(fmt.Sprintf("Respuesta de %s:%d/%s %v", ip, puerto, endpoint, resData))
	return resData.Mensaje
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

	respuesta := structs.Respuesta {
		Mensaje: fmt.Sprint(http.StatusOK),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
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

func Decode(line string) (interface{}) {
    parts := strings.Fields(line) // Divide por espacios
    if len(parts) == 0 {
		slog.Error("línea vacía")
        return nil
    }

    cmd := parts[0]
    params := parts[1:]

    instType, ok := instructionMap[cmd]
    if (!ok) {
		slog.Error(fmt.Sprintf("comando desconocido: %s", cmd))
        return nil
    }

    switch instType {
    case structs.INST_NOOP:
        if len(params) != 0 { 
			slog.Error("NOOP no espera parámetros")
			return nil
		}
        return structs.NoopInstruction{}

    case structs.INST_WRITE:
        if len(params) != 2 { 
			slog.Error("WRITE espera 2 parámetros (Dirección, Datos)")
			return nil
		}
        addr, err := strconv.Atoi(params[0])
        if err != nil { 
			slog.Error(fmt.Sprintf("parámetro Dirección inválido para WRITE: %v", err))
			return nil
		}
        return structs.WriteInstruction{Address: addr, Data: params[1]}

     case structs.INST_READ:
         if len(params) != 2 { 
			slog.Error("READ espera 2 parámetros (Dirección, Tamaño)")
			return nil
		}
         addr, err := strconv.Atoi(params[0])
         if err != nil { 
			slog.Error(fmt.Sprintf("parámetro Dirección inválido para READ: %v", err))
			return nil
		}
         size, err := strconv.Atoi(params[1])
         if err != nil { 
			slog.Error(fmt.Sprintf("parámetro Tamaño inválido para READ: %v", err)) 	
			return nil
		}
         return structs.ReadInstruction{Address: addr, Size: size}

     case structs.INST_GOTO:
         if len(params) != 1 { 
			slog.Error("GOTO espera 1 parámetro (Valor)")
			return nil
		}
         target, err := strconv.Atoi(params[0])
         if err != nil { 
			slog.Error(fmt.Sprintf("parámetro Valor inválido para GOTO: %v", err)) 
			return nil
		}
         return structs.GotoInstruction{TargetAddress: target}

	
    case structs.INST_IO:
        if len(params) != 2 {
            slog.Error("IO espera 2 parámetros (Duración, Nombre)")
            return nil
        }
        duration, err := strconv.Atoi(params[0])
        if err != nil {
            slog.Error(fmt.Sprintf("parámetro Duración inválido para IO: %v", err))
            return nil
        }
        return structs.IoInstruction{Duration: duration, Nombre: params[1]}

    case structs.INST_INIT_PROC:
        if len(params) != 2 {
            slog.Error("INIT_PROC espera 2 parámetros (NombreProceso, TamañoMemoria)")
            return nil
        }
        memorySize, err := strconv.Atoi(params[1])
        if err != nil {
            slog.Error(fmt.Sprintf("parámetro TamañoMemoria inválido para INIT_PROC: %v", err))
            return nil
        }
        return structs.InitProcInstruction{ProcessName: params[0], MemorySize: memorySize}

     case structs.INST_DUMP_MEMORY:
         if len(params) != 0 { 
			slog.Error("DUMP_MEMORY no espera parámetros")
			return nil
		}
         return structs.DumpMemoryInstruction{}

     case structs.INST_EXIT:
        if len(params) != 0 { 
			slog.Error("EXIT no espera parámetros")
			return nil
		}
        return structs.ExitInstruction{}

    default:
		slog.Error(fmt.Sprintf("parsing no implementado para: %s", cmd))
        return nil
    }
}

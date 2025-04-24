package structs

type Interfaz struct {
	IP		string
	Puerto	int
}

type PeticionIO struct {
	Nombre		string
	Interfaz	Interfaz
}

type PeticionKernel struct {
	PID            	uint
	NombreIfaz		string
	SuspensionTime 	int
}
package common

const (
	MAX_RECURSIION_DEPTH = uint8(4)
	MAX_VM_INSTANCES     = uint64(2048)
)

const (
	SUB_PROCESS = iota
	CONTAINER_ID
	ELEMENT_ID
	UUID
)

var TotalSubProcesses uint64

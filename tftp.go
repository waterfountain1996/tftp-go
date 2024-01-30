package tftp

type opcode uint16

const (
	opRRQ opcode = iota + 1
	opWRQ
	opDATA
	opACK
	opERROR
	opOACK
)

package tftp

type opcode uint16

func (op opcode) String() string {
	switch op {
	case opRRQ:
		return "RRQ"
	case opWRQ:
		return "WRQ"
	case opDATA:
		return "DATA"
	case opACK:
		return "ACK"
	case opERROR:
		return "ERROR"
	default:
		return "OACK"
	}
}

const (
	opRRQ opcode = iota + 1
	opWRQ
	opDATA
	opACK
	opERROR
	opOACK
)

const (
	errUndefined uint16 = iota
	errNotFound
	errPermission
	errDiskFull
	errIllegalOp
	errUnknownTID
	errAlreadyExists
	errNoSuchUser
)

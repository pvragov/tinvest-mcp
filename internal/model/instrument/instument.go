package instrument

type Type int

//go:generate go run github.com/dmarkham/enumer -type=Type -text -json -yaml -transform=lower -trimprefix=Type -output=type_enum.go
const (
	TypeUnknown Type = iota
	TypeShare
	TypeBond
	TypeCurrency
	TypeETF
)

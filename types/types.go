package orm

type Type interface {
	String() string
}

type typ int

const (
	Int8 typ = iota
	Int16
	Int32
	Int64
	Float32
	Float64
	String
	Time
	Bool
)

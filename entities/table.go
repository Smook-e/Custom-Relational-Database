package entities

// Data type identifiers
const (
	TypeTinyInt uint8 = 0 //1 byte
	TypeSmallInt uint8 = 1 // 2 bytes
	TypeInt     uint8 = 2 // 4 bytes
	TypeBigInt  uint8 = 3 // 8 bytes
	TypeVarChar uint8 = 4 
)

// Constraint identifiers 
const (
	ConstraintNone       uint8 = 0
	ConstraintPrimaryKey uint8 = 1 << 0 // 1
	ConstraintNotNull    uint8 = 1 << 1 // 2
	ConstraintUnique     uint8 = 1 << 2 // 4
)

type Table struct {

	RootIndex int
	Columns []Column
}

type Column struct {
	Name string
	DataType uint8
	Contstraints uint8
	Size uint8
}


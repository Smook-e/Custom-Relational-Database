package entities

// Data type identifiers
const (
	TypeSmallInt uint8 = 1 
	TypeInt     uint8 = 2 
	TypeBigInt  uint8 = 3 
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

	rootIndex int
	columns []Column
}

type Column struct {
	name string
	dataType uint8
	contstraints uint8
	size uint8
}
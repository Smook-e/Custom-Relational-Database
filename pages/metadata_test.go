package pages

import (
	"path/filepath"
	"testing"

	"github.com/Smook-e/Custom-Relational-Database/entities"

	// "fmt"
	"os"
	"reflect"
)

func newEmptyDatabase(t *testing.T, dbPath string) *entities.Database {
	t.Helper()

	filep, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("failed to open database file: %v", err)
	}
	db := &entities.Database{
		File:      filep,
		Tables:    make(map[string]*entities.Table),
		FreePages: make([]entities.FreePage, 0),
		TotalPages: uint32(0),
	}
	return db
}

func TestReadAndWriteMetadataPage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "storage-test.db")
	db := newEmptyDatabase(t, dbPath)
	tableNames := []string{"users", "products", "orders"}
	// Create a sample table
	for _, name := range tableNames {
		table := &entities.Table{
				Name: name,
				Columns: make([]entities.Column, 0),
				Indexes:      make(map[string]uint32),
				ForeignKeys:  make(map[string]entities.ForeignKeyReference),
			}
		db.Tables[table.Name] = table
		cols := []struct{
			Name        string
			DataType    string
			Constraints []string
			DefaultValue any
		}{
			{"id", "serial", []string{"primarykey"}, nil},
			{"name", "varchar", []string{"notnull", "default"}, "anonymous"},
			{"email", "varchar(60)", []string{"unique"}, nil},
			{"age", "int", []string{"default"}, int32(18)},
			{"user_id", "int", []string{"notnull"}, nil},
		}
		for _, col := range cols {
			Column := entities.Column{
				Name:         col.Name,
			}
			dataType,Size, err := entities.GetDataTypeAndSize(col.DataType)
			if err != nil {
				t.Fatalf("failed to get data type and size for column %s: %v", col.Name, err)
			}
			constraints , err := entities.GetConstraint(col.Constraints)
			if err != nil {
				t.Fatalf("failed to get constraint for column %s: %v", col.Name, err)
			}
			Column.DataType = dataType
			Column.Size = Size
			Column.Constraints = constraints
			if col.DefaultValue != nil {
				Column.Default = col.DefaultValue
			}
			table.Columns = append(table.Columns, Column)
		}
		table.Indexes["id"] = 1
		table.ForeignKeys["user_id"] = entities.ForeignKeyReference{
			ReferencedTableName: "users",
			ReferencedColumnIndex: 0,
		}
	}
	db.FreePages = append(db.FreePages, entities.FreePage{PageID: 2, FreeSpace: 100})
	db.FreePages = append(db.FreePages, entities.FreePage{PageID: 3, FreeSpace: 200})
	err :=WriteMetaPage(db)
	if err != nil {
		t.Fatalf("failed to write metadata page: %v", err)
	}
	db.File.Close()
	readDB := newEmptyDatabase(t, dbPath)
	err = ReadMetaPage(readDB)
	if err != nil {
		t.Fatalf("failed to read metadata page: %v", err)
	}
	// Validate metadata matches what was written
	if readDB.TotalPages != db.TotalPages {
		t.Fatalf("TotalPages mismatch: got %d want %d", readDB.TotalPages, db.TotalPages)
	}
	if len(readDB.FreePages) != len(db.FreePages) {
		t.Fatalf("FreePages length mismatch: got %d want %d", len(readDB.FreePages), len(db.FreePages))
	}
	for i, fp := range db.FreePages {
		if readDB.FreePages[i] != fp {
			t.Fatalf("FreePages[%d] mismatch: got %+v want %+v", i, readDB.FreePages[i], fp)
		}
	}

	for _, name := range tableNames {
		orig, ok := db.Tables[name]
		if !ok {
			t.Fatalf("original table %s missing", name)
		}
		rd, ok := readDB.Tables[name]
		if !ok {
			t.Fatalf("read table %s missing", name)
		}

		if orig.Name != rd.Name {
			t.Fatalf("table name mismatch: got %s want %s", rd.Name, orig.Name)
		}

		if len(orig.Columns) != len(rd.Columns) {
			t.Fatalf("columns count for %s mismatch: got %d want %d", name, len(rd.Columns), len(orig.Columns))
		}

		for i := range orig.Columns {
			oc := orig.Columns[i]
			rc := rd.Columns[i]
			if oc.Name != rc.Name || oc.DataType != rc.DataType || oc.Size != rc.Size || oc.Constraints != rc.Constraints {
				t.Fatalf("column %d in table %s mismatch: got %+v want %+v", i, name, rc, oc)
			}
			if oc.HasConstraint(entities.ConstraintDefault) {
				if !reflect.DeepEqual(oc.Default, rc.Default) {
					t.Fatalf("default value mismatch for column %s.%s: got %+v want %+v", name, oc.Name, rc.Default, oc.Default)
				}
			} else {
				if rc.Default != nil {
					t.Fatalf("expected no default for column %s.%s but got %+v", name, oc.Name, rc.Default)
				}
			}
		}

		if !reflect.DeepEqual(orig.Indexes, rd.Indexes) {
			t.Fatalf("indexes mismatch for table %s: got %+v want %+v", name, rd.Indexes, orig.Indexes)
		}
		if len(orig.ForeignKeys) != len(rd.ForeignKeys) {
			t.Fatalf("foreign keys count mismatch for table %s: got %d want %d", name, len(rd.ForeignKeys), len(orig.ForeignKeys))
		}
		for k, v := range orig.ForeignKeys {
			rv, ok := rd.ForeignKeys[k]
			if !ok {
				t.Fatalf("foreign key %s missing in read table %s", k, name)
			}
			if v.ReferencedTableName != rv.ReferencedTableName || v.ReferencedColumnIndex != rv.ReferencedColumnIndex {
				t.Fatalf("foreign key %s mismatch in table %s: got %+v want %+v", k, name, rv, v)
			}
		}
	}

}
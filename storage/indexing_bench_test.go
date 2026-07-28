package storage
import (
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"fmt"
	
)
func BenchmarkIndexSearch(b *testing.B) {
	// Initialize the storage engine and insert 100,000 BigInt keys into the index
	filename := "database.bin"
	engine, err := InitializeStorageEngine(filename)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer engine.Bp.File.Close()

	root := engine.TestIndexInsertRoot(0)// insert 1,000,000 BigInt keys into the index
	b.ResetTimer()
	//search 
	for  b.Loop() {
		key, err := engine.db.Serialize(int64(999999), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key 999999 ","Error:", err)
			return
		}
	}
	engine.db.File.Truncate(0)
}
func BenchmarkLinearSearch(b *testing.B) {
	// Initialize the storage engine
	filename := "database.bin"
	engine, err := InitializeStorageEngine(filename)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer engine.Bp.File.Close()

	root := engine.TestIndexInsertRoot(0)// insert 1,000,000 BigInt keys into the index
	b.ResetTimer()
	//search 
	for b.Loop()  {
		key, err := engine.db.Serialize(int64(999999), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.LinearSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key 999999 ","Error:", err)
			return
		}
	}
	engine.db.File.Truncate(0)
}
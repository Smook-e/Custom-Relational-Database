package storage
import (
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"fmt"
	
)
func BenchmarkIndexSearch(b *testing.B) {
	// Initialize the storage engine and insert 100,000 BigInt keys into the index
	engine, err := InitializeStorageEngine("database.bin")
	if err != nil {
		fmt.Println("Error initializing storage engine:", err)
		return
	}
	defer engine.Bp.File.Close()
	engine.TestWriteandReadDatabase()
	engine, err = InitializeStorageEngine("database.bin")
	if err != nil {
		fmt.Print(err)
		return
	}

	engine.TestIndexInsertRoot(0)// insert 1000,000 BigInt keys into the index
	var root uint32 = 346 // root page id after inserting 1000,000 BigInt keys
	b.ResetTimer()
	//search 
	for i := 0; i <= 1000000; i++ {
		key, err := engine.db.Serialize(int64(i % 1000000), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			continue
		}
	}
}
func BenchmarkLinearSearch(b *testing.B) {
	// Initialize the storage engine and insert 100,000 BigInt keys into the index
	engine, err := InitializeStorageEngine("database.bin")
	if err != nil {
		fmt.Println("Error initializing storage engine:", err)
		return
	}
	defer engine.Bp.File.Close()
	engine.TestWriteandReadDatabase()
	engine, err = InitializeStorageEngine("database.bin")
	if err != nil {
		fmt.Print(err)
		return
	}

	engine.TestIndexInsertRoot(0)// insert 1000,000 BigInt keys into the index
	var root uint32 = 346 // root page id after inserting 1000,000 BigInt keys
	b.ResetTimer()
	//search 
	for i := 0; i <= 1000000; i++ {
		key, err := engine.db.Serialize(int64(i % 1000000), entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.LinearSearch(root, key, entities.TypeBigInt)
		if err != nil {
			fmt.Println("Error searching for key:", i, "Error:", err)
			return
		}
	}
}
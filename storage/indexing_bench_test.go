package storage
import (
	"testing"
	"github.com/Smook-e/Custom-Relational-Database/entities"
	"fmt"
	
)


/*
This file contains benchmark tests for the B+Tree index implementation in the database.
It provides functions to benchmark the performance of searching for keys in the index using both the B+Tree search and a linear search approach.
You can find the results of the benchmark tests in the readme.md file.
*/
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
		key, err := engine.db.Serialize(int64(999999), &entities.Column{DataType: entities.TypeBigInt, Size: 8})
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.IndexSearch(root, key, &entities.Column{DataType: entities.TypeBigInt, Size: 8})
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
		key, err := engine.db.Serialize(int64(999999), &entities.Column{DataType: entities.TypeBigInt, Size: 8})
		if err != nil {
			fmt.Println("Error serializing key:", err)
			return
		}
		_, _, err = engine.LinearTree(root, key, &entities.Column{DataType: entities.TypeBigInt, Size: 8})
		if err != nil {
			fmt.Println("Error searching for key 999999 ","Error:", err)
			return
		}
	}
	engine.db.File.Truncate(0)
}
package bleve

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/kawai-network/y/paths"
	"github.com/kawai-network/y/jarvis/db"
)

// BleveDB actually uses DuckDB now, but we keep the name for compatibility
type BleveDB struct {
	db      *sql.DB
	Hash    string
	Session string
}

var (
	bleveDB *BleveDB
	once    sync.Once
)

func getDBPath() string {
	os.MkdirAll(paths.Jarvis(), 0755)
	return paths.JarvisAddressBookDB()
}

func getHashPath() string {
	return paths.JarvisAddressBookHash()
}

func getDataFromDefaultFile() (result map[string]string, hash string) {
	result = make(map[string]string)
	dir := paths.Jarvis()
	file := path.Join(dir, "addresses.json")
	var timestamp int64
	fi, err := os.Lstat(file)
	if err != nil {
		fmt.Printf("reading addresses from ~/addresses.json failed: %s. Ignored.\n", err)
		return map[string]string{}, fmt.Sprintf("%d", timestamp)
	}
	// if the file is a symlink
	if fi.Mode()&os.ModeSymlink != 0 {
		file, err = os.Readlink(file)
		if err != nil {
			fmt.Printf("reading addresses from ~/addresses.json failed: %s. Ignored.\n", err)
			return map[string]string{}, fmt.Sprintf("%d", timestamp)
		}
	}
	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("reading addresses from ~/addresses.json failed: %s. Ignored.\n", err)
		return map[string]string{}, fmt.Sprintf("%d", timestamp)
	}

	info, err := os.Stat(file)
	if err != nil {
		fmt.Printf("reading addresses from ~/addresses.json failed: %s. Ignored.\n", err)
		return map[string]string{}, fmt.Sprintf("%d", timestamp)
	}
	timestamp += info.ModTime().UnixNano()

	err = json.Unmarshal(content, &result)
	if err != nil {
		fmt.Printf("reading addresses from ~/addresses.json failed: %s. Ignored.\n", err)
		return map[string]string{}, fmt.Sprintf("%d", timestamp)
	}

	content, err = os.ReadFile(path.Join(dir, "secrets.json"))
	if err == nil {
		secret := map[string]string{}
		err = json.Unmarshal(content, &secret)
		if err == nil {
			for addr, name := range secret {
				result[addr] = name
			}
		}
	}
	info, err = os.Stat(path.Join(dir, "secrets.json"))
	if err == nil {
		timestamp += info.ModTime().UnixNano()
	}

	for addr, tokenName := range db.TOKENS {
		result[addr] = tokenName
	}
	return result, fmt.Sprintf("%d", timestamp)
}

func NewBleveDB() (*BleveDB, error) {
	var resError error
	once.Do(func() {
		bleveDB = &BleveDB{}
		// Open DuckDB
		d, err := sql.Open("duckdb", getDBPath())
		if err != nil {
			resError = fmt.Errorf("failed to open duckdb: %w", err)
			return
		}
		bleveDB.db = d

		// Load extensions
		if _, err := d.Exec("INSTALL fts; LOAD fts;"); err != nil {
			// Just log, might be already installed/loaded or not needed for basic usage,
			// but critical for FTS. Fails if offline and not cached.
			fmt.Printf("Warning: Failed to install/load FTS extension: %v\n", err)
		}

		resError = loadData(bleveDB)
	})
	return bleveDB, resError
}

func loadData(bdb *BleveDB) error {
	addresses, newHash := getDataFromDefaultFile()

	// Check hash to avoid rebuilding if data hasn't changed
	hashBytes, _ := os.ReadFile(getHashPath())
	if string(hashBytes) == newHash {
		// Verify table exists
		var count int
		if err := bdb.db.QueryRow("SELECT count(*) FROM information_schema.tables WHERE table_name = 'addresses'").Scan(&count); err == nil && count > 0 {
			bdb.Hash = newHash
			return nil
		}
	}

	// Rebuild Table
	if _, err := bdb.db.Exec("DROP TABLE IF EXISTS addresses"); err != nil {
		return err
	}
	// We use address as ID for FTS index
	if _, err := bdb.db.Exec("CREATE TABLE addresses (address TEXT PRIMARY KEY, description TEXT)"); err != nil {
		return err
	}

	// Batch Insert
	// Since DuckDB local is fast, single tx is usually enough
	tx, err := bdb.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO addresses (address, description) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for addr, desc := range addresses {
		if _, err := stmt.Exec(addr, desc); err != nil {
			// Continue or fail? addresses map key is unique, but let's be safe
			fmt.Printf("Failed to insert %s: %v\n", addr, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	// Create FTS Index
	// Indexing 'address' and 'description'. 'address' is the key column.
	if _, err := bdb.db.Exec("PRAGMA create_fts_index('addresses', 'address', 'address', 'description')"); err != nil {
		return fmt.Errorf("failed to create fts index: %w", err)
	}

	// Save Hash
	if err := os.WriteFile(getHashPath(), []byte(newHash), 0644); err != nil {
		fmt.Printf("Warning: failed to write hash file: %v\n", err)
	}
	bdb.Hash = newHash
	return nil
}

func (bleveDB *BleveDB) Persist() error {
	// No-op for DuckDB as it persists to file automatically
	// But we might want to force checkpoint?
	return nil
}

func (bleveDB *BleveDB) Search(input string) ([]AddressDesc, []int) {
	if bleveDB.db == nil {
		return []AddressDesc{}, []int{}
	}

	// Hybrid Search:
	// 1. Full Text Search (BM25) on Address & Description
	// 2. Levenshtein Distance on Description (for fuzzy name match)
	// Score calculation attempts to mimic Bleve's score scale roughly

	query := `
	WITH results AS (
		SELECT 
			address, 
			description, 
			fts_main_addresses.match_bm25(address, ?) * 10 AS score_ft 
		FROM addresses 
		WHERE score_ft IS NOT NULL
		
		UNION ALL
		
		SELECT 
			address, 
			description,
			(1.0 / (levenshtein(description, ?) + 0.1)) * 5 AS score_fuzzy
		FROM addresses
		WHERE levenshtein(description, ?) <= 2
	)
	SELECT address, description, SUM(score_ft) as total_score
	FROM results
	GROUP BY address, description
	ORDER BY total_score DESC
	LIMIT 20
	`

	rows, err := bleveDB.db.Query(query, input, input, input)
	if err != nil {
		fmt.Printf("Address db search failed: %s\n", err)
		return []AddressDesc{}, []int{}
	}
	defer rows.Close()

	results := []AddressDesc{}
	resultScores := []int{}

	for rows.Next() {
		var addr, desc string
		var score float64
		if err := rows.Scan(&addr, &desc, &score); err != nil {
			continue
		}

		// Normalize score to int as expected by caller (Bleve used quite large ints)
		intScore := int(score * 100000)

		results = append(results, AddressDesc{
			Address: addr,
			Desc:    desc,
		})
		resultScores = append(resultScores, intScore)
	}

	return results, resultScores
}

// remove unused functions to avoid linter warnings if any
// indexAddresses was removed

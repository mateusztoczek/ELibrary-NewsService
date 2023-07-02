package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"news/config"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateNewsTable(t *testing.T) {
	configData, err := ioutil.ReadFile("../config/testConfig.json")
	if err != nil {
		t.Fatal("failed to read config file:", err)
	}

	var testConfig config.Config
	err = json.Unmarshal(configData, &testConfig)
	if err != nil {
		t.Fatal("failed to parse config file:", err)
	}

	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require",
		testConfig.Host, testConfig.Port, testConfig.User, testConfig.Password, testConfig.DBName)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		t.Fatal("failed to connect to the database:", err)
	}
	defer db.Close()

	err = CreateNewsTable(db, testConfig)

	assert.NoError(t, err)

	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = $1
		AND table_name = $2
	)`
	var exists bool
	err = db.QueryRow(query, testConfig.SchemaName, testConfig.TableName).Scan(&exists)
	if err != nil {
		t.Fatal("failed to query the database:", err)
	}
	assert.True(t, exists)
}

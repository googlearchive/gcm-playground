// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	"./lib"
	"./models"
	"./routes"
)

// What port should the server run on
const port string = "4260"

var (
	// The name of the database to connect to
	databaseName = "data.db"

	// Print logging
	debug = true
)

func InitDb() gorm.DB {
	// Database connection
	// Run `go run migrate.go` before running the server, or after making
	// changes to the models.
	db, err := gorm.Open("sqlite3", databaseName)
	if err != nil {
		log.Fatal(err)
	}
	db.DB()
	db.AutoMigrate(&models.Client{})
	db.LogMode(debug) // Helps with debugging

	// Database Connections for each package
	models.InitializeDb(db)
	routes.InitializeDb(db)

	return db
}

func main() {
	InitDb()

	// Start the server
	log.Println(fmt.Sprintf("Started, serving at port %v", port))
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), lib.Handler())
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}

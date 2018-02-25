// Golang RESTful JSON API server,
// is a Backend for working with TREE data structure,
// uses NESTED SET MODEL to work with database(MySQL)
// link: http://mikehillyer.com/articles/managing-hierarchical-data-in-mysql/ ,
// supports Cross-Origin Resource Sharing (CORS).

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // database driver
	"github.com/rs/cors"               // package to handle CORS requests
)

// Node holds info for a node in the tree.
type Node struct {
	ID         int
	Title      string
	ChildNodes []Node
}

// rNodeData holds parameters provided by Request.Body,
// needed for adding nodes to the database.
type rNodeData struct {
	ParentID int
	Title    string
}

// db - declaration of database object.
var db *sql.DB

// Declare the variable to hold errors.
var err error

// dbInit sets up database connection,
// creates database instance,
// creates table to work with,
// inserts ROOT node.
func dbInit() {
	fmt.Println("-------ACCESSING MySQL SERVER-------")
	// Parameters, user have to provide to access the database
	var (
		username string
		password string
	)
	fmt.Print("Enter user_name: ")
	fmt.Scanln(&username)
	fmt.Print("Enter password: ")
	fmt.Scanln(&password)
	params := fmt.Sprintf("%s:%s@/", username, password)
	db, err = sql.Open("mysql", params)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	// Set up special mode to work with NESTED SET MODEL.
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("CREATE DATABASE IF NOT EXISTS treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS NodesTable (
		    ID INT NOT NULL AUTO_INCREMENT,
		    Title VARCHAR(20) NOT NULL,
		    lft INT NOT NULL,
		    rgt INT NOT NULL,
			Avatar VARCHAR(255) NOT NULL DEFAULT 'react.png',
		    PRIMARY KEY (ID)
			)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("TRUNCATE TABLE NodesTable")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("INSERT INTO NodesTable (Title, lft, rgt) VALUES ('ROOT', 1, 2)")
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

// dbGetDescendants takes a Node and populates the field ChildNodes(i.e array)
// with all descendants(i.e. nodes) down the hierarchy.
func dbGetDescendants(node *Node) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	stmtdbGetDescendants, err := tx.Prepare(
		`SELECT node.ID, node.Title, (COUNT(parent.Title) - (sub_tree.depth + 1)) AS depth
        FROM NodesTable AS node,
        NodesTable AS parent,
        NodesTable AS sub_parent,
        (
                SELECT node.Title, (COUNT(parent.Title) - 1) AS depth
                FROM NodesTable AS node,
                        NodesTable AS parent
                WHERE node.lft BETWEEN parent.lft AND parent.rgt
                        AND node.ID = ?
                GROUP BY node.Title
                ORDER BY node.lft
            ) AS sub_tree
                    WHERE node.lft BETWEEN parent.lft AND parent.rgt
        AND node.lft BETWEEN sub_parent.lft AND sub_parent.rgt
        AND sub_parent.Title = sub_tree.Title
                GROUP BY node.Title
                HAVING depth = 1
                ORDER BY node.lft`)
	if err != nil {
		return err
	}
	defer stmtdbGetDescendants.Close()

	rows, err := stmtdbGetDescendants.Query(node.ID)
	if err == sql.ErrNoRows {
		return errors.New("no descendants for current node in database")
	} else if err != nil {
		return err
	}

	for rows.Next() {
		var Child Node
		// depth - redundant parameter just to scan
		// from database query and forget about it.
		var depth int

		err = rows.Scan(&Child.ID, &Child.Title, &depth)
		if err != nil {
			return err
		}
		node.ChildNodes = append(node.ChildNodes, Child)
	}
	if err = rows.Err(); err != nil {
		return err
	} else {
		rows.Close()
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	// Check if the current node has children
	// and set off func dbGetDescendants down the tree.
	if node.ChildNodes != nil {
		for i := range node.ChildNodes {
			dbGetDescendants(&node.ChildNodes[i])
		}
	}
	return nil
}

// dbGetTree returns the whole tree,
// within the ROOT node.
func dbGetTree() (*Node, error) {
	// Tree variable contains the tree
	var Tree Node

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	// Fetch the ROOT node (i.e. ID=1) from database.
	err = tx.QueryRow("SELECT ID, Title FROM NodesTable WHERE ID = ?", 1).Scan(&Tree.ID, &Tree.Title)
	if err == sql.ErrNoRows {
		return nil, errors.New("no data for root node in database")
	} else if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	// Get all descendants for the ROOT node.
	err = dbGetDescendants(&Tree)
	if err != nil {
		return nil, err
	}
	return &Tree, nil
}

// duplicateCounter indicates the number of a node's copy
// by checking node's "Title",
// added invisibly to "Title" if node has duplicates in database.
var duplicateCounter int

// dbAddNewNode adds a new node to the database,
// requires ID of the node(i.e. parent), which will be
// joined with a child with provided "Title" and "Avatar".
func dbAddNewNode(parentID int, Title string, Avatar string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	// Check for a duplicate "Title" in database and handle if find one.
	var duplTitle string
	err = tx.QueryRow("SELECT Title FROM NodesTable WHERE Title = ?", Title).Scan(&duplTitle)
	if duplTitle == Title && err == nil {
		Title += string(duplicateCounter)
		duplicateCounter += 1
	}

	// Check for avatar error
	if Avatar == "" {
		Avatar = "react.png" // default
	}

	_, err = tx.Exec("LOCK TABLE NodesTable WRITE")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`SELECT @myPoint := rgt - 1 FROM NodesTable
	WHERE ID = ?`, parentID)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE NodesTable SET rgt = rgt + 2 WHERE rgt > @myPoint")
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE NodesTable SET lft = lft + 2 WHERE lft > @myPoint")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO NodesTable(Title, lft, rgt, Avatar)
	VALUES(?, @myPoint + 1, @myPoint + 2, ?)`, Title, Avatar)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UNLOCK TABLES")
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// dbDeleteNode removes node from database with a given ID.
func dbDeleteNode(id int) error {
	if id == 1 {
		return errors.New("must not delete root node")
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("LOCK TABLE NodesTable WRITE")
	if err != nil {
		return err
	}
	_, err = tx.Exec(`SELECT @myLeft := lft, @myRight := rgt, @myWidth := rgt - lft + 1
	FROM NodesTable
	WHERE ID = ?`, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM NodesTable WHERE lft BETWEEN @myLeft AND @myRight")
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE NodesTable SET rgt = rgt - @myWidth WHERE rgt > @myRight")
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE NodesTable SET lft = lft - @myWidth WHERE lft > @myRight")
	if err != nil {
		return err
	}
	_, err = tx.Exec("UNLOCK TABLES")
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// getTreeHandler returns Response containing the whole tree in the JSON format.
func getTreeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Fetch the tree from database
	Tree, err := dbGetTree()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(&Tree)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
}

// Counts imports of avatar files.
var fileCounter int

// Prepare RegExp to validate Title for addNodeHandler.
var validTitle = regexp.MustCompile("^([a-zA-Z0-9]+)$")

// addNodeHandler adds a new node to the database,
// returns Response, containing the whole tree in the JSON format.
func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	var fileName string
	// Check if an avatar was uploaded.
	if fileStatus := r.FormValue("filestatus"); fileStatus == "true" {
		// if true, fetch file from Request and save to ./avatars,
		// using unique name for the file.
		file, info, err := r.FormFile("uploadfile")
		defer file.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		fileType := info.Header["Content-Type"][0][6:]
		fileName = "avatar" + strconv.Itoa(fileCounter) + "." + fileType
		fileCounter += 1
		f, err := os.OpenFile("avatars/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}

	// Fetch JSON from Request.
	jsonData := r.FormValue("jsonData")
	b := []byte(jsonData)
	var data rNodeData
	err = json.Unmarshal(b, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Validate Title.
	m := validTitle.FindStringSubmatch(data.Title)
	if m == nil || data.Title == "" {
		data.Title = "Node"
	}

	// Add node to the database
	err = dbAddNewNode(data.ParentID, data.Title, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	// Redirect to return the whole tree within Response.
	getTreeHandler(w, r)
}

// Prepare RegExp to validate URL path for deleteNodeHandler.
var validPathDelete = regexp.MustCompile("^/deleteNode/ID=([0-9]+)$")

// deleteNodeHandler removes the node with the given ID from the database,
// returns Response containing the whole tree in the JSON format.
func deleteNodeHandler(w http.ResponseWriter, r *http.Request) {
	// Validate Request URL.
	m := validPathDelete.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	ID, _ := strconv.Atoi(m[1])

	// Remove the node from database.
	err = dbDeleteNode(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	// Redirect to return the whole tree within Response.
	getTreeHandler(w, r)
}

// Prepare RegExp to validate URL path for getAvatarHandler.
var validPathGet = regexp.MustCompile("^/getAvatar/ID=([0-9]+)$")

// getAvatarHandler returns avatar's source file with Response
// for the given node ID.
func getAvatarHandler(w http.ResponseWriter, r *http.Request) {
	// Validate Request URL.
	m := validPathGet.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	ID := m[1] // Fetch node's ID.

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
	}
	_, err = tx.Exec("USE treestorage")
	if err != nil {
		fmt.Println(err)
	}
	_, err = tx.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		fmt.Println(err)
	}

	// Fetch name of the avatar's source file from database.
	var fileName string
	err = tx.QueryRow("SELECT Avatar from NodesTable WHERE ID = ?", ID).Scan(&fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Path to the avatar's source file.
	avatarPath := "avatars/" + fileName

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
	}

	// Return proper Response Header.
	content, err := ioutil.ReadFile(avatarPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	fileContentType := http.DetectContentType(content)
	w.Header().Set("Content-Type", fileContentType)

	// Return avatar within Response
	Openfile, err := os.Open(avatarPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	defer Openfile.Close()
	io.Copy(w, Openfile)
}

func main() {
	dbInit()
	defer db.Close()

	// Testing database.
	err = dbAddNewNode(1, "NODE 1", "")
	if err != nil {
		log.Fatal(err)
	}

	// Start HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/getTree", getTreeHandler)
	mux.HandleFunc("/addNode", addNodeHandler)
	mux.HandleFunc("/deleteNode/", deleteNodeHandler)
	mux.HandleFunc("/getAvatar/", getAvatarHandler)
	// Add a handler to CORS requests.
	handler := cors.Default().Handler(mux)
	log.Println("Server has started -> http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

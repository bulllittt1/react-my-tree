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
	"log"
	"net/http"

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
// needed for actions: adding or deleting nodes from the database.
type rNodeData struct {
	ID       int
	ParentID int
	Title    string
}

// db - declaration of database object.
var db *sql.DB

// Declare the variable to hold errors.
var err error

// dbGetDescendants takes a Node type and populates the field ChildNodes(i.e array)
// with all descendants(i.e. nodes) down the tree.
func dbGetDescendants(db *sql.DB, node *Node) error {
	stmtdbGetDescendants, err := db.Prepare(
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
		// depth - redundant parameter just to fetch
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
	// Check if the current node has children
	// and set off dbGetDescendants function down the tree.
	if node.ChildNodes != nil {
		for i := range node.ChildNodes {
			dbGetDescendants(db, &node.ChildNodes[i])
		}
	}
	return nil
}

// dbGetTree returns the whole tree,
// firstly, it fetches the ROOT node from database
// and then it attaches all descendants to the ROOT.
func dbGetTree(db *sql.DB) (*Node, error) {
	// Tree variable contains the tree
	var Tree Node

	stmtGetRoot, err := db.Prepare("SELECT ID, Title FROM NodesTable WHERE ID = ?")
	if err != nil {
		return nil, err
	}
	// stmtGetRoot fetches the ROOT node (i.e. ID=1) from database.
	err = stmtGetRoot.QueryRow(1).Scan(&Tree.ID, &Tree.Title)
	if err == sql.ErrNoRows {
		return nil, errors.New("no data for root node in database")
	} else if err != nil {
		return nil, err
	}
	// Get all descendants for the ROOT node.
	err = dbGetDescendants(db, &Tree)
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
// joined with a child wiht provided "Title".
func dbAddNewNode(db *sql.DB, parentID int, Title string) error {
	// Check for a duplicate "Title" in database and handle if find one.
	var duplTitle string
	_ = db.QueryRow("SELECT Title FROM NodesTable WHERE Title = ?", Title).Scan(&duplTitle)
	if duplTitle == Title {
		Title += string(duplicateCounter)
		duplicateCounter += 1
	}

	tx, err := db.Begin()
	if err != nil {
		return err
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
	_, err = tx.Exec(`INSERT INTO NodesTable(Title, lft, rgt)
	VALUES(?, @myPoint + 1, @myPoint + 2)`, Title)
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

// dbDeleteNode removes node from database with a specified ID.
func dbDeleteNode(db *sql.DB, id int) error {
	if id == 1 {
		return errors.New("must not delete root node")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
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

// getTreeHandler handles HTTP "GET" method Request,
// returns Response containing the whole tree in the JSON format.
func getTreeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Fetch the tree from database
	Tree, err := dbGetTree(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(&Tree)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
}

// addNodeHandler handles HTTP "POST" method Request,
// containing JSON with required parameters to add a new node to the database,
// returns Response, containing the whole tree in the JSON format.
func addNodeHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	// data - container for parameters of the new node.
	var data rNodeData
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	defer r.Body.Close()
	if data.Title == "" {
		data.Title = "No title"
	}

	// Add a new node to the database
	err = dbAddNewNode(db, data.ParentID, data.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	// Redirect to return the whole tree with Response.
	getTreeHandler(w, r)
}

// deleteNodeHandler handles HTTP "POST" method Request,
// containing JSON with required parameters to delete a node from the database,
// returns Response containing the whole tree in the JSON format.
func deleteNodeHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	// data - container for parameters of the node to delete.
	var data rNodeData
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
		return
	}
	defer r.Body.Close()
	// Remove the node from database
	err = dbDeleteNode(db, data.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	// Redirect to return the whole tree with Response.
	getTreeHandler(w, r)
}

// dbInit sets up database connection,
// creates database instance,
// creates table to work with,
// inserts ROOT node.
func dbInit() {
	fmt.Println("-------ACCESSING MySQL SERVER-------")
	// Parameters user have to provide to access the database
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
	// Set up special mode to work with NESTED SET MODEL.
	_, err = db.Exec(`SET sql_mode=(SELECT REPLACE(@@sql_mode,'ONLY_FULL_GROUP_BY',''))`)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("CREATE DATABASE IF NOT EXISTS TreeStorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec("USE TreeStorage")
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(
		`CREATE TABLE IF NOT EXISTS NodesTable (
		    ID INT NOT NULL AUTO_INCREMENT,
		    Title VARCHAR(20) NOT NULL,
		    lft INT NOT NULL,
		    rgt INT NOT NULL,
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

func main() {
	dbInit()
	defer db.Close()

	// Testing database.
	err = dbAddNewNode(db, 1, "NODE 1")
	if err != nil {
		log.Fatal(err)
	}

	// Start HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/getTree", getTreeHandler)
	mux.HandleFunc("/addNode", addNodeHandler)
	mux.HandleFunc("/deleteNode", deleteNodeHandler)
	// Add a handler to CORS requests.
	handler := cors.Default().Handler(mux)
	log.Println("Server has started -> http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

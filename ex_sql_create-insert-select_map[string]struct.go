package main
import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"strconv"
	// "github.com/Masterminds/squirrel"
	// "github.com/jmoiron/sqlx"
)
type Article struct {
	/* var section1 string
	var section2 string */
	Sections map[string]string
	NodeID uint64
	Pagerank float64
}
func main() {
	db, err := sql.Open(
		"mysql",
		"root:root@tcp(localhost:3311)/sample")
	if err != nil { panic(err) }
	defer db.Close()

	// var article map[string]Article
	var sections1 map[string]string
	var sections2 map[string]string
	sections1 = make(map[string]string)
	sections2 = make(map[string]string)
	sections2["foo2"] = "bar2"
	sections1["foo1"] = "bar1"
	article1 := Article{
		Sections: sections1,
		NodeID: 1,
		Pagerank: 0.75,
	}	
	article2 := Article{
		Sections: sections2,
		NodeID: 2,
		Pagerank: 0.75,
	}
	article := map[string]Article{
		"article_1": article1,
		"article_2": article2,
	}
	fmt.Println(article)
	err = SqlInsert(db, article)
	if err != nil { panic(err) }
	/* rows, err := db.Query("SELECT * FROM title");
	if err != nil { panic(err) }
	defer rows.Close() */
	articles, err := MyRowScanner("article_1", db)
	if err != nil { panic(err) }
	for key, _ := range articles {
		fmt.Println(key, articles[key].NodeID, articles[key].Pagerank)
		for section, val := range articles[key].Sections {
			fmt.Println(section, val)
		}
	}
	articles, err = MyRowScanner("article_2", db)
	if err != nil { panic(err) }
	for key, _ := range articles {
		fmt.Println(key, articles[key].NodeID, articles[key].Pagerank)
		for section, val := range articles[key].Sections {
			fmt.Println(section, val)
		}
	}
}
func MyRowScanner(table string, db *sql.DB) (articles map[string]Article, err error) {
	rows, err := db.Query("SELECT * FROM " + table);
	if err != nil { panic(err) }
	defer rows.Close()

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]sql.RawBytes, count)
	valuePtrs := make([]interface{}, count)
	
	articles = make(map[string]Article)
	articles[table] = Article{}
	// articles[table].Sections I= make(map[string]string)
	var i int
	i = -1
	for rows.Next() {
		i++
		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)

		// var tmp Article
		// tmp.Sections = make(map[string]string)
		var tmp = articles[table]
		tmp.Sections = make(map[string]string)
		for i, _ := range columns {
			switch columns[i] {
			case "NodeID":
				if values[i] != nil {
					// fmt.Println("case \"NodeID\": columns[i]:\t", columns[i], values[i])
					tmp.NodeID, err = strconv.ParseUint(string(values[i]), 10, 32)
					// articles[table].NodeID, err = strconv.ParseUint(string(values[i]), 10, 32)
					if err != nil { return nil, err }
				}
			case "Pagerank":
				if values[i] != nil {
					// fmt.Println("case \"Pagerank\": columns[i]:\t", columns[i], values[i])
					tmp.Pagerank, err = strconv.ParseFloat(string(values[i]), 32)
					// articles[table].Pagerank, err = strconv.ParseFloat(string(values[i]), 32)
					if err != nil { return nil, err }
				}
			default:
				if values[i] != nil {
					// fmt.Println("default: columns[i]:\t", columns[i], values[i])
					tmp.Sections[columns[i]] = string(values[i])
					// articles[table].Sections[columns[i]] = string(values[i])
				}
			}
		}
		articles[table] = tmp
	}
	fmt.Println()
	
	return articles, nil
}

// TODO: Exec stmt check err's to make values (items.*) correspond to *=?
func SqlInsert(db *sql.DB, articles map[string]Article) (err error) {
	for title, items := range articles {
		// sections := make([]string, len(items.Sections))
		_, err := db.Exec("CREATE TABLE  " + title + "(NodeID int, Pagerank float)")
		if err != nil { return err }
		_, err = db.Exec("INSERT INTO " + title + " (NodeID, Pagerank) VALUES (?, ?)", items.NodeID, items.Pagerank)
		if err != nil { return err }
		for sectionTitle, sectionBody := range items.Sections {
			// _, err = db.Exec("ALTER TABLE ? ADD ? text", title, sectionTitle)
			_, err = db.Exec("ALTER TABLE " + title + " ADD COLUMN " + sectionTitle + " text")
			if err != nil { return err }
			// _, err = db.Exec("INSERT INTO ($1) VALUES ($2)", title, sectionBody,)
			_, err = db.Exec("INSERT INTO " + title + " (" + sectionTitle + ") VALUES (?)", sectionBody)
			if err != nil { return err }
		}
	}
	return nil
}

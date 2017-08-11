package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

const VERSION = "0.0.1"

var (
	database string
	bucket   string

	key   string
	value string

	add    bool
	delete bool
	read   bool
	update bool

	print bool

	noop    bool
	verbose bool
	debug   bool

	version bool

	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	table  = tablewriter.NewWriter(os.Stdout)
)

func init() {
	log.SetFlags(0)
	flag.StringVar(
		&database, "f", os.Getenv("MJOLNIR_DB"),
		"The path to the boltdb file\n\tDefault: MJOLNIR_DB environment variable",
	)
	flag.StringVar(
		&bucket, "b", os.Getenv("MJOLNIR_DB_BUCKET"),
		"The name of the boltdb bucket\n\tDefault: MJOLNIR_DB_BUCKET environment variable",
	)

	flag.StringVar(&key, "key", "", "The key to add,delete,read,update.")
	flag.StringVar(&value, "value", "", "The value to add,update.")

	flag.BoolVar(&add, "a", false, "Add a key/value to the boltdb database")
	flag.BoolVar(&delete, "d", false, "Delete the key from the boltdb database")
	flag.BoolVar(&read, "r", false, "Read the key from the boltdb database")
	flag.BoolVar(&update, "u", false, "Update the key/value in the boltdb database")
	flag.BoolVar(&print, "p", false, "Print entire bucket to stdout.")

	flag.BoolVar(&noop, "n", false, "Prints operation to stdout instead of acting on the boltdb database.")
	flag.BoolVar(&verbose, "v", false, "Prints verbose info messages to stdout.")
	flag.BoolVar(&debug, "D", false, "Prints debug messages to stdout.")

	flag.BoolVar(&version, "V", false, "Prints the version number of Mjolnir")
}

func main() {
	flag.Parse()

	if version {
		log.Println(VERSION)
		os.Exit(0)
	}

	if len(database) <= 0 {
		flag.Usage()
		log.Fatalln(yellow("[ ERROR ]: BoltDB path not set\nPlease pass the [ -f /path/to/bolt.db ] argument or set the MJOLNIR_DB environment variable"))
	}

	db, err := bolt.Open(database, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mjolnir := New(db, bucket)

	// Setup Markdown Table
	setupTable()

	if !print && !add && !delete && !read && !update {
		flag.Usage()
		log.Fatalln(yellow("[ ERROR ]: You must choose an operation. Add, Delete, Read, Update, Print"))
	}

	if (!print || read || delete) && len(key) <= 0 {
		flag.Usage()
		log.Fatalln(yellow("[ ERROR ]: Key not set.\nPlease pass the [ -key key_name ] arguments"))
	}

	if (add || update) && len(value) <= 0 {
		flag.Usage()
		log.Fatalln(yellow("[ ERROR ]: You must pass a value when adding or updating an entry in the boltdb database.\nPlease pass the [ -value '{ \"foo\" : \"bar\" }' ] arugment"))
	}

	if print && len(bucket) <= 0 {
		mjolnir.PrintBuckets()
		os.Exit(0)
	} else if !print && len(bucket) <= 0 {
		flag.Usage()
		log.Fatalln(yellow("[ ERROR ]: Bucket name not set\nPlease pass the [ -b bucket_name ] argument or set the MJOLNIR_DB_BUCKET environment variable"))
	} else if print && len(bucket) >= 1 && len(key) <= 0 {
		mjolnir.Print()
		os.Exit(0)
	} else if print && len(key) >= 1 && len(bucket) >= 1 {
		mjolnir.Print()
		os.Exit(0)
	}

	if add {
		err := mjolnir.Create(key, value)
		if err != nil {
			log.Fatalln(err)
		}
		mjolnir.PrintRecord()
	}

	if delete {
		err := mjolnir.Delete(key)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if read {
		mjolnir.PrintRecord()
	}

	if update {
		err := mjolnir.Update(key, value)
		if err != nil {
			log.Fatalln(err)
		}
		mjolnir.PrintRecord()
	}
}

// Mjolnir - defines the mjolnir type
type Mjolnir struct {
	*bolt.DB

	bucket string
}

func (m *Mjolnir) PrintRecord() {
	// Configure Table
	table.SetHeader([]string{"Key", "Value"})

	result, err := m.Read(key)
	if err != nil {
		log.Fatalln(err)
	}
	str, err := json.Marshal(result)
	if err != nil {
		log.Fatalln("[ ERROR ]: Encoding JSON failed")
	}

	row := []string{fmt.Sprintf("%s", key), fmt.Sprintf("%s", str)}
	table.Append(row)

	// Render Table
	table.Render()
}

// PrintBuckets - prints a list of all buckets.
func (m *Mjolnir) PrintBuckets() {
	// Configure Table
	table.SetHeader([]string{"Bucket Name"})

	recordCount := 0
	err := m.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			table.Append([]string{string(name)})
			recordCount++
			return nil
		})
	})
	if err != nil {
		log.Fatalln(err)
	}
	// Render Table
	table.Render()
	fmt.Printf(" total buckets: %d\n",
		recordCount,
	)

}

// Print - will print an entire bucket to the screen or an individual entry
func (m *Mjolnir) Print() error {
	return m.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(m.bucket))
		if err != nil {
			log.Fatalln(err)
		}

		table.SetHeader([]string{"Key", "Value"})
		recordCount := 0
		// Iterate Bucket b
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			row := []string{fmt.Sprintf("%s", k), fmt.Sprintf("%s", v)}
			table.Append(row)
			recordCount++
		}
		table.Render()
		fmt.Printf(" database: %s | bucket: %s | total records: %d\n",
			database,
			m.bucket,
			recordCount,
		)
		return nil
	})
	return nil
}

// Create - creates a new key in bolt db
func (m *Mjolnir) Create(key string, value interface{}) error {
	err := m.putDBKey(key, value)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("[ INFO ]: creating key \"%s\" with value \"%v\" succeeded.\n", key, value)
	}
	return nil
}

// Read - reads a key in bolt db
func (m *Mjolnir) Read(key string) (interface{}, error) {
	var v interface{}
	var err error

	if v, err = m.getDBKey(key); err != nil {
		return nil, err
	}

	if v == nil {
		return nil, fmt.Errorf("[ ERROR ]: key \"%s\" was not found in bucket \"%s\"", key, m.bucket)
	}

	if verbose {
		fmt.Printf("[ INFO ]: key \"%s\" found.\n", key)
	}

	return v, err
}

// Update - updates a key in bolt db
func (m *Mjolnir) Update(key string, value interface{}) error {
	if err := m.Delete(key); err != nil {
		return err
	}

	if err := m.Create(key, value); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("[ INFO ]: key \"%s\" updated.\n", key)
	}

	return nil
}

// Delete - deletes a key in bolt db
func (m *Mjolnir) Delete(key string) error {
	if err := m.deleteDBKey(key); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("[ INFO ]: key \"%s\" deleted.\n", key)
	}

	return nil
}

// New - returns a new bolt db
func New(db *bolt.DB, bucket string) *Mjolnir {
	m := &Mjolnir{
		DB:     db,
		bucket: bucket,
	}
	return m
}

// ------------------------------------------------------------------
// Private Methods
// ------------------------------------------------------------------
func (m *Mjolnir) deleteDBKey(key string) error {
	return m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(m.bucket))
		if err != nil {
			return err
		}

		err = bucket.Delete([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})
}

func (m *Mjolnir) getDBKey(key string) (interface{}, error) {
	var value interface{}
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(m.bucket))
		if err != nil {
			return err
		}

		val := bucket.Get([]byte(key))
		if val == nil {
			return err
		}

		err = json.Unmarshal(val, &value)
		if err != nil {
			return err
		}

		return nil
	})
	return value, err
}

func (m *Mjolnir) putDBKey(key string, v interface{}) error {
	err := m.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(m.bucket))
		if err != nil {
			return err
		}

		// Check if the value is a json string
		if func(s string) bool {
			var js interface{}
			return json.Unmarshal([]byte(s), &js) == nil
		}(v.(string)) {
			err = bucket.Put([]byte(key), []byte(v.(string)))
			if err != nil {
				return err
			}
			return nil
		}

		data, err := json.Marshal(v)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), []byte(data))
		if err != nil {
			return err
		}

		return nil

	})
	return err
}

func setupTable() {
	// Configure Table
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
}

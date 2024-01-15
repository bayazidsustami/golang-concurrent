package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bayazidsustami/golang-concurrent/database"
	_ "github.com/go-sql-driver/mysql"
)

const (
	csvFile     = "assets/majestic_million.csv"
	totalWorker = 80
)

var dataHeaders = make([]string, 0)

func main() {
	start := time.Now()

	db, err := database.OpenDBConnection()
	if err != nil {
		log.Fatal(err.Error())
	}

	csvReader, csvFile, err := openCsvFile()
	if err != nil {
		log.Fatal(err.Error())
	}
	defer csvFile.Close()

	jobs := make(chan []any)
	wg := new(sync.WaitGroup)

	go dispatchWorker(db, jobs, wg)
	readCsvFilePerLineThenSendToWorker(csvReader, jobs, wg)

	wg.Wait()

	duration := time.Since(start)
	fmt.Println("done in", int(math.Ceil(duration.Seconds())), "seconds")
}

func dispatchWorker(db *sql.DB, jobs <-chan []any, wg *sync.WaitGroup) {
	for workerIndex := 0; workerIndex < totalWorker; workerIndex++ {
		go func(workerIndex int, db *sql.DB, jobs <-chan []any, wg *sync.WaitGroup) {
			counter := 0
			for job := range jobs {
				doTheJob(workerIndex, counter, db, job)
				wg.Done()
				counter++
			}
		}(workerIndex, db, jobs, wg)
	}
}

func readCsvFilePerLineThenSendToWorker(csvReader *csv.Reader, jobs chan<- []any, wg *sync.WaitGroup) {
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}

		if len(dataHeaders) == 0 {
			dataHeaders = row
			continue
		}

		rowOrdered := make([]any, 0)
		for _, value := range row {
			rowOrdered = append(rowOrdered, value)
		}

		wg.Add(1)
		jobs <- rowOrdered
	}

	close(jobs)
}

func openCsvFile() (*csv.Reader, *os.File, error) {
	log.Println("=> open csv file")

	file, err := os.Open(csvFile)
	if err != nil {
		return nil, nil, err
	}

	reader := csv.NewReader(file)
	return reader, file, nil
}

func doTheJob(workerIndex, counter int, db *sql.DB, values []any) {
	for {
		var outerError error
		func(outerError *error) {
			defer func() {
				if err := recover(); err != nil {
					*outerError = fmt.Errorf("%v", err)
				}
			}()

			conn, err := db.Conn(context.Background())
			if err != nil {
				log.Fatal(err.Error())
			}
			query := fmt.Sprintf("INSERT INTO domain (%s) VALUES (%s)",
				strings.Join(dataHeaders, ","),
				strings.Join(generateQuestionsMark(len(dataHeaders)), ","),
			)

			_, err = conn.ExecContext(context.Background(), query, values...)
			if err != nil {
				log.Fatal(err.Error())
			}

			err = conn.Close()
			if err != nil {
				log.Fatal(err.Error())
			}
		}(&outerError)
		if outerError == nil {
			break
		}
	}

	if counter%100 == 0 {
		log.Println("=> worker", workerIndex, "inserted", counter, "data")
	}
}

func generateQuestionsMark(n int) []string {
	s := make([]string, 0)
	for i := 0; i < n; i++ {
		s = append(s, "?")
	}
	return s
}

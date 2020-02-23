package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type Worker struct {
	*sync.WaitGroup
}

func makeRequestAndCalculateCount(url string) (int, error) {
	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Fatalln(err.Error())
			return
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return bytes.Count(body, []byte("Go")), nil
}

func (w *Worker) DoWork(chForUrls <- chan string, chForResults chan <- int) {
	go func() {
		for url := range chForUrls {
			count, err := makeRequestAndCalculateCount(url)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return
			}

			chForResults <- count

			fmt.Printf("Count for %s: %d\n", url, count)
		}

		w.Done()
		return
	}()
}

func calculateResult(chForResults <- chan int, chTotalResult chan <- int) {
	go func() {
		totalCnt := 0
		for count := range chForResults {
			totalCnt += count
		}
		chTotalResult <- totalCnt

		return
	}()
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	existsWorkersCnt := 0
	maxWorkersCnt := 5
	wg := sync.WaitGroup{}

	chForUrls := make(chan string)
	chForResults := make(chan int)
	chTotalResult := make(chan int)

	calculateResult(chForResults, chTotalResult)

	for scanner.Scan() {
		if existsWorkersCnt < maxWorkersCnt {
			existsWorkersCnt++

			wg.Add(1)
			worker := &Worker{&wg}
			worker.DoWork(chForUrls, chForResults)
		}

		chForUrls <- scanner.Text()
	}

	close(chForUrls)
	wg.Wait()

	close(chForResults)

	fmt.Printf("Total: %d\n", <-chTotalResult)
}
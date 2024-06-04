package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ronsor/rwkv-tokenizer-go"
)

var (
	inputPath      = flag.String("input", "wikipedia_simple.jsonl", "Input data file")
	inputFormat    = flag.String("input-format", "json", "Input data format (json, nullsep)")
	inputTextField = flag.String("input-field", "text", "Text field key for JSON format")

	statsInterval = flag.Duration("stats-interval", 5*time.Second, "Interval for printing current stats")
)

var (
	stats struct {
		tokens int64
		bytes  int64
		start  time.Time
		end    time.Time
	}

	quitFlag bool
)

func readInputInner(out chan string) {
	switch *inputFormat {
	case "nullsep":
		f, err := os.Open(*inputPath)
		if err != nil {
			log.Fatal("could not open data file:", err)
		}
		defer f.Close()

		bf := bufio.NewReader(f)
		for {
			doc, err := bf.ReadString('\x00')
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println("failed to read data file:", err)
				break
			}
			out <- doc
		}
	case "json":
		var m map[string]string
		f, err := os.Open(*inputPath)

		if err != nil {
			log.Fatal("could not open data file:", err)
		}
		defer f.Close()

		bf := bufio.NewReader(f)
		for {
			line, err := bf.ReadBytes('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println("failed to read data file:", err)
				break
			}
			err = json.Unmarshal(line, &m)
			if err != nil {
				log.Println("failed to parse data file:", err)
				break
			}
			doc, ok := m[*inputTextField]
			if !ok {
				log.Println("missing text field key")
				break
			}
			out <- doc
		}
	}

	close(out)
}

func readInput() chan string {
	ch := make(chan string)
	go readInputInner(ch)
	return ch
}

func printStats(full bool) {
	now := time.Now()
	if !stats.end.IsZero() {
		now = stats.end
	}
	fmt.Printf("\rTokens: %10d | Bytes: %12d | Elapsed: %20s", stats.tokens, stats.bytes, now.Sub(stats.start).String())
	if full {
		timeDiff := float64(now.Sub(stats.start)/time.Millisecond)/1000
		fmt.Printf(
			"\nElapsed sec: %10.04f\nBytes/token: %10.02f\nTokens/sec:  %10.02f\nBytes/sec:   %10.02f\n",
			timeDiff,
			float64(stats.bytes)/float64(stats.tokens),
			float64(stats.tokens)/timeDiff,
			float64(stats.bytes)/timeDiff,
		)
	}
}

func statReporter() {
	i := 0
	for {
		time.Sleep(*statsInterval)
		printStats(i == 0)
		i = (i + 1) % 5
		quitFlag = false
	}
}

func signalHandler(ch chan os.Signal) {
	for sig := range ch {
		switch sig {
		case os.Interrupt:
			printStats(true)
			if !quitFlag {
				fmt.Println("*** Use ^C again to quit ***")
				quitFlag = true
			} else {
				log.Fatal("interrupted")
			}
		}
	}
}

func main() {
	flag.Parse()

	tokenizer := rwkvtkn.NewWorldTokenizer()
	dataset := readInput()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go signalHandler(ch)

	stats.start = time.Now()
	go statReporter()
	for doc := range dataset {
		tokens, err := tokenizer.EncodeString(doc)
		if err != nil {
			log.Fatal("tokenizer error:", err)
		}
		stats.tokens += int64(len(tokens))
		stats.bytes += int64(len(doc))
	}
	stats.end = time.Now()

	fmt.Println("\n--- final stats ---")
	printStats(true)
	fmt.Println("\n--- ----------- ---")
}

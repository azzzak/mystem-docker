package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// default timeout for request in milliseconds
var timeout time.Duration = 1000

var (
	stdin  io.WriteCloser
	stdout io.ReadCloser
)

var errEmptyString = errors.New("Empty string")

var version = "unknown"

func main() {
	fmt.Printf("mystem-docker %s\n", version)
	args := []string{"-i", "--eng-gr", "--weight", "--format=json"}

	homonyms, exists := os.LookupEnv("HOMONYMS_DETECTION")
	if exists && isTrue(homonyms) {
		args = append(args, "-d")
	}

	glue, exists := os.LookupEnv("GLUE_GRAMMEMES")
	if exists && isTrue(glue) {
		args = append(args, "-g")
	}

	dict, exists := os.LookupEnv("USER_DICT")
	if exists {
		file := fmt.Sprintf("/stem/dict/%s", dict)
		if !isDictExist(file) {
			log.Fatalf("Can't find user dictionary \"%s\"\n", dict)
		}

		d := fmt.Sprintf("--fixlist=%s", file)
		args = append(args, d)
	}

	timeLimit, exists := os.LookupEnv("TIMEOUT")
	if exists {
		v, err := strconv.Atoi(timeLimit)
		if err != nil {
			log.Fatalln("Timeout must be integer")
		}
		timeout = time.Duration(v)
	}

	var err error
	if stdin, stdout, err = runMystem(args); err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	defer stdin.Close()
	defer stdout.Close()

	http.HandleFunc("/mystem", limit(listen))
	http.ListenAndServe(":8080", nil)
}

func runMystem(args []string) (io.WriteCloser, io.ReadCloser, error) {
	cmd := exec.Command("./mystem", args...)

	var err error
	if stdin, err = cmd.StdinPipe(); err != nil {
		return nil, nil, err
	}

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return nil, nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}

	return stdin, stdout, nil
}

func isTrue(s string) bool {
	s = strings.ToLower(s)
	if s == "true" || s == "yes" {
		return true
	}
	return false
}

func isDictExist(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func process(ctx context.Context, text string) (string, error) {
	if len(text) == 0 {
		return "", errEmptyString
	}

	resChan := make(chan string)
	errChan := make(chan error)

	go func() {
		defer close(resChan)
		defer close(errChan)

		fmt.Fprintln(stdin, text)
		buf := bufio.NewReader(stdout)

		str, err := buf.ReadString('\n')
		if err != nil {
			errChan <- err
		}
		resChan <- strings.TrimSuffix(str, "\n")
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-resChan:
		return r, nil
	case err := <-errChan:
		return "", err
	}
}

func limit(f http.HandlerFunc) http.HandlerFunc {
	sema := make(chan struct{}, 1)
	return func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		f(w, r)
		<-sema
	}
}

func listen(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	err := r.ParseForm()
	if err != nil {
		log.Printf("Error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s := r.Form.Get("text")

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, timeout*time.Millisecond)
	p, err := process(ctx, s)
	cancel()

	if err != nil {
		switch err {
		case errEmptyString:
			http.Error(w, err.Error(), http.StatusNoContent)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Printf("Error: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(p))
}

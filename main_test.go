package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_process(t *testing.T) {
	type args struct {
		ctx  context.Context
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Try word",
			args:    args{context.Background(), "слон"},
			want:    genOutput("слон"),
			wantErr: false,
		},
		{
			name:    "Try sentence",
			args:    args{context.Background(), "Съешь еще этих мягких французских булок"},
			want:    genOutput("Съешь еще этих мягких французских булок"),
			wantErr: false,
		},
		{
			name:    "Try empty string",
			args:    args{context.Background(), ""},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			stdin = &helperWriteCloser{
				io.Writer(b),
			}

			fmt.Fprintln(stdin, tt.args.text)

			buf := bufio.NewReader(b)
			str, err := buf.ReadString('\n')
			if err != nil {
				t.Log(err)
			}

			s := fmt.Sprintln(genOutput(str))
			bout := bytes.NewBufferString(s)
			stdout = ioutil.NopCloser(bout)

			got, err := process(tt.args.ctx, tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listen(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    string
		wantErr bool
	}{
		{
			name:    "POST слон",
			text:    "слон",
			want:    genOutput("слон"),
			wantErr: false,
		},
		{
			name:    "POST sentence",
			text:    "Съешь еще этих мягких французских булок",
			want:    genOutput("Съешь еще этих мягких французских булок"),
			wantErr: false,
		},
		{
			name:    "POST empty string",
			text:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			stdin = &helperWriteCloser{
				io.Writer(b),
			}
			fmt.Fprintln(stdin, tt.text)

			buf := bufio.NewReader(b)
			str, err := buf.ReadString('\n')
			if err != nil {
				t.Log(err)
			}

			res := fmt.Sprintln(genOutput(str))
			bout := bytes.NewBufferString(res)
			stdout = ioutil.NopCloser(bout)

			data := url.Values{}
			data.Set("text", tt.text)

			req, err := http.NewRequest("POST", "/stem", strings.NewReader(data.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(limit(listen))
			handler.ServeHTTP(rr, req)
			status := rr.Code

			if tt.wantErr && status != http.StatusOK {
				return
			}

			if status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			if rr.Body.String() != tt.want {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.want)
			}
		})
	}
}

func Test_isTrue(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true",
			args: args{s: "true"},
			want: true,
		},
		{
			name: "yes",
			args: args{s: "yes"},
			want: true,
		},
		{
			name: "TrUe",
			args: args{s: "TrUe"},
			want: true,
		},
		{
			name: "YES",
			args: args{s: "YES"},
			want: true,
		},
		{
			name: "false",
			args: args{s: "false"},
			want: false,
		},
		{
			name: "no",
			args: args{s: "no"},
			want: false,
		},
		{
			name: "Empty string",
			args: args{s: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTrue(tt.args.s); got != tt.want {
				t.Errorf("isTrue() = %v, want %v", got, tt.want)
			}
		})
	}
}

type helperWriteCloser struct {
	io.Writer
}

func (mwc *helperWriteCloser) Close() error {
	return nil
}

func genOutput(sentence string) string {
	if len(sentence) == 0 {
		return ""
	}

	sentence = strings.TrimSuffix(sentence, "\n")
	words := strings.Split(sentence, " ")

	var res []string
	for _, v := range words {
		s := fmt.Sprintf("{\"analysis\":[{\"lex\":\"%s\",\"wt\":1,\"gr\":\"\"}],\"text\":\"%s\"}", v, v)
		res = append(res, s)
	}
	return fmt.Sprintf("[%s]", strings.Join(res, ","))
}

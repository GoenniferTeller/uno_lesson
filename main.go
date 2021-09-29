package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func getGoEnv() []string {
	cmd := exec.Command("go", "env")
	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	ret := []string{}
	for _, e := range strings.Split(string(output), "\n") {
		ret = append(ret, strings.Replace(e, "\"", "", -1))
	}
	return ret
}

func getGoRoot() string {
	goEnv := getGoEnv()
	for _, env := range goEnv {
		if strings.Contains(env, "GOROOT") {
			return strings.Split(env, "=")[1]
		}
	}

	log.Fatal("Unable to find GOROOT")
	return ""
}

func copyGoWasmJsFile() {
	goRoot := getGoRoot()
	goWasmFile := filepath.Join(goRoot, "misc/wasm/wasm_exec.js")
	target := filepath.Join("wasm_app", "wasm_exec.js")
	if err := Copy(goWasmFile, target); err != nil {
		log.Fatal(err)
	}
}

var wasmBuildArgs = []string{
	"GOOS=js",
	"GOARCH=wasm",
}

func build() {
	env := getGoEnv()
	for _, arg := range wasmBuildArgs {
		env = append(env, arg)
	}

	cmd := exec.Command("go", "build", "-o", "lib.wasm")
	cmd.Dir = "wasm_app"
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal(err)
	}
}

func startServer() {
	http.Handle("/", http.FileServer(http.Dir("wasm_app")))

	port := "8000"
	log.Printf("Serving HTTP on port: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	log.Println("Building..")
	copyGoWasmJsFile()
	build()

	log.Println("Starting server..")
	startServer()
}

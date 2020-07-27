package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/samuel/go-thrift/parser"
)

func (g *GoGenerator) writeFile(out *bytes.Buffer, outfile string) {
	outBytes := out.Bytes()
	if g.Format {
		var err error
		outBytes, err = format.Source(outBytes)
		if err != nil {
			fmt.Println(out.String())
			g.error(err)
		}
	}

	fi, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(2)
	}
	if _, err := fi.Write(outBytes); err != nil {
		fi.Close()
		g.error(err)
	}
	fi.Close()
}

func (g *GoGenerator) generateRoutes(out io.Writer) error {
	g.write(out, "// This file is automatically generated. Do not modify.\n")
	g.write(out, "\npackage endpoints\n")

	return nil
}

func (g *GoGenerator) generateServices(out io.Writer, path string, thrift *parser.Thrift) error {
	packageName := g.Packages[path].Name

	g.write(out, "// This file is automatically generated. Do not modify.\n")
	g.write(out, "\npackage endpoints\n")
	g.write(out, "\nimport (\n")
	g.write(out, "\"encoding/json\"\n")
	g.write(out, "\"errors\"\n")
	g.write(out, "\"%s\"\n", packageName)
	g.write(out, ")\n")

	for _, k := range sortedKeys(thrift.Services) {
		svc := thrift.Services[k]
		svcName := camelCase(svc.Name)
		methodNames := sortedKeys(svc.Methods)

		g.write(out, "\nvar %sEndpointByMethod = map[string]endpoint{\n", svcName)

		for _, k := range methodNames {
			method := svc.Methods[k]
			methodName := camelCase(method.Name)
			g.write(out, "\"%s\": %s%s,\n", method.Name, svcName, methodName)
		}

		g.write(out, "}\n")
	}

	for _, k := range sortedKeys(thrift.Services) {
		svc := thrift.Services[k]
		svcName := camelCase(svc.Name)
		methodNames := sortedKeys(svc.Methods)

		for _, k := range methodNames {
			method := svc.Methods[k]
			methodName := camelCase(method.Name)

			g.write(out, "\nfunc %s%s(client %s.RPCClient, requestBytes []byte) (interface{}, error) {\n", svcName, methodName, packageName)
			g.write(out, "req := &%s.%s%sRequest{}\n", packageName, svcName, methodName)
			g.write(out, "res := &%s.%s%sResponse{}\n", packageName, svcName, methodName)
			g.write(out, "if err := json.Unmarshal(requestBytes, req); err != nil {\n")
			g.write(out, "return nil, err\n")
			g.write(out, "}\n")
			g.write(out, "if err := client.Call(\"%s\", req, res); err != nil {\n", method.Name)
			g.write(out, "return nil, err\n")
			g.write(out, "}\n")
			g.write(out, "return res, nil\n")
			g.write(out, "}\n")
		}

		g.write(out, "\nfunc %sService(client %s.RPCClient, requestBytes []byte, method string) (interface{}, error) {\n", svcName, packageName)
		g.write(out, "endpoint, ok := %sEndpointByMethod[method]\n", svcName)
		g.write(out, "if !ok {\n")
		g.write(out, "return nil, errors.New(\"Unsupported method\")\n")
		g.write(out, "}\n")

		g.write(out, "\nreturn endpoint(client, requestBytes)\n")
		g.write(out, "}\n")
	}

	return nil
}

func (g *GoGenerator) generateEndpoints(outPath string) {
	pkgpath := filepath.Join(outPath, "endpoints")

	if err := os.MkdirAll(pkgpath, 0755); err != nil {
		g.error(err)
	}

	out := &bytes.Buffer{}
	outfile := filepath.Join(pkgpath, "routes.go")

	g.generateRoutes(out)
	g.writeFile(out, outfile)

	for path, th := range g.ThriftFiles {
		if len(th.Services) > 0 {
			filename := strings.ToLower(filepath.Base(path))
			for i := len(filename) - 1; i >= 0; i-- {
				if filename[i] == '.' {
					filename = filename[:i]
				}
			}
			filename += ".go"
			outfile := filepath.Join(pkgpath, filename)

			out := &bytes.Buffer{}
			if err := g.generateServices(out, path, th); err != nil {
				g.error(err)
			}

			g.writeFile(out, outfile)
		}
	}
}

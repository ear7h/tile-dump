/*
 * can be run as like so:
 *	psql -d natural_earth -t -c "select 'CoastLine', ST_AsText(wkb_geometry) from ne_10m_coastline limit 1 offset 1;" | go run wkt_conv.go >> natural_earth.go
 */
package main

import (
	"fmt"
	"os"
	"strings"
	"io/ioutil"
)

func geomType(typ string) string {
	if strings.HasPrefix(typ, "MULTI") {
		return "Multi" + geomType(typ[len("MULTI"):])
	}

	switch typ {
		case "POINT":
			return "Point"

		case "LINESTRING":
			return "LineString"

		case "POLYGON":
			return "Polygon"

		case "GEOMETRYCOLLECTION":
			return "Collection"
	}
	panic("not found " + typ)
}

func parseGeom(w io.Writer, byt byte) int {
	n = bytes.Index(byt, "(")
	typ := bytes.TrimSpace(byt[:n])
	typ = geomType(string(typ))



	if typ == "Collection" {
		fmt.Fprintf(w, "geom.Collection{")

		for {
			n = parseGeom(w, byt[n+1:])
		}

		w.Write([]byte("}")
	}

	fmt.Fprintf(w, "&geom.%s", typ)

	buf := make([]byte, 1)
	for _, v := range byt[n:]{
		switch v {
		case ',':
			w.Write([]byte("},{"))
		case ' ':
			w.Write([]byte(", "))
		case '(':
			w.Write([]byte("{"))
		case ')':
			w.Write([]byte("}"))
		default:
			buf[0] = v
			w.Write(buf)
		}
	}

	// return fmt.Printf('geom.%s{%s}', typ, str)
}

func findClosingParen(str string) int {
	n := 0
	for i, v := range str {
		if v == '(' {
			n++
		} else if v == ')' {
			n--
		}

		if n < 0 {
			return i
		}
	}

	return -1
}

func main() {
	byt, _ := ioutil.ReadAll(os.Stdin)

	str := string(byt)
	n := strings.Index(str, "|")
	name := strings.Replace(str[:n], " ", "", -1)

	str = str[n+1:]
	fmt.Printf("var %s = geom.%s{%s}\n", name, typ, str)
}

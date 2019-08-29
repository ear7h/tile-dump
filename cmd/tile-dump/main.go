package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ear7h/tile-dump/log"
	"github.com/ear7h/tile-dump/register"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola/atlas"
	"github.com/go-spatial/tegola/basic"
	"github.com/go-spatial/tegola/config"
	"github.com/go-spatial/tegola/dict"
	"github.com/go-spatial/tegola/provider"
)

var (
	conf config.Config
)

func initConfig(configFile string, cacheRequired bool) (err error) {
	log.Infof("Loading config file: %v", configFile)
	if conf, err = config.Load(configFile); err != nil {
		return err
	}
	if err = conf.Validate(); err != nil {
		return err
	}

	// init our providers
	// but first convert []env.Map -> []dict.Dicter
	provArr := make([]dict.Dicter, len(conf.Providers))
	for i := range provArr {
		provArr[i] = conf.Providers[i]
	}

	providers, err := register.Providers(provArr)
	if err != nil {
		return fmt.Errorf("could not register providers: %v", err)
	}

	// init our maps
	if err = register.Maps(nil, conf.Maps, providers); err != nil {
		return fmt.Errorf("could not register maps: %v", err)
	}
	if len(conf.Cache) == 0 && cacheRequired {
		return fmt.Errorf("No cache defined in config, please check your config (%v).", configFile)
	}

	return nil
}

func main() {
	config := os.Args[1]
	tile, err := Format{
		X:   1,
		Y:   2,
		Z:   0,
		Sep: "/",
	}.ParseTile(os.Args[2])
	if err != nil {
		panic(err)
	}

	fmt.Println("dumping: ", config, tile)

	err = initConfig(config, false)
	if err != nil {
		panic(err)
	}

	fmt.Println("len maps: ", len(atlas.AllMaps()))

	m := atlas.AllMaps()[0]
	ret := geom.Collection{}

	for _, layer := range m.Layers {
		ptile := provider.NewTile(tile.Z,
			tile.X,
			tile.Y,
			uint(m.TileBuffer),
			uint(m.SRID))

		ctx := context.Background()
		err = layer.Provider.TileFeatures(ctx, layer.ProviderLayerName, ptile, func(f *provider.Feature) error {
			geo, err := basic.ToWebMercator(f.SRID, f.Geometry)
			if err != nil {
				return err
			}
			g, ok := geo.(geom.Collection)
			if ok {
				ret = append(ret, g...)
			} else {
				ret = append(ret, geo)
			}

			return nil
		})

		if err != nil {
			panic(err)
		}
	}

	/*
	str, err := wkt.Encode(ret)
	if err != nil {
		panic(err)
	}
	*/

	name := m.Name + "_" + strings.Replace(os.Args[2], "/", "_", -1)

	f, err := os.OpenFile(
		name + ".go",
		os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
		0666)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(f, `package testing
import (
	"github.com/go-spatial/geom"
)

`)
	fmt.Fprintf(f, "var _%s = ", name)
	WriteGoGeom(f, ret)
	// n, err := io.Copy(f, strings.NewReader(str))
	// fmt.Println("copy: ", n, err)
	f.Close()

}

func WriteGoGeom(w io.Writer, geo geom.Geometry) error {
	col, ok := geo.(geom.Collection)
	if ok {
		fmt.Fprintf(w, "geom.Collection{")
		for _, v := range col {
			err := WriteGoGeom(w, v)
			if err != nil {
				return err
			}
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "}")
		return nil
	} else {
		fmt.Fprintf(w, "%T", geo)
		return writeGoGeom(w, geo)
	}
}

func writeGoGeom(w io.Writer, geo geom.Geometry) error {
	switch g := geo.(type) {
	case geom.Point:
		fmt.Fprintf(w, "{%v, %v}", g[0], g[1])
	case geom.LineString:
		fmt.Fprintf(w, "{")
		for _, v := range g {
			writeGoGeom(w, geom.Point(v))
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, "}")
	case geom.MultiPoint:
		fmt.Fprintf(w, "{")
		for _, v := range g {
			writeGoGeom(w, geom.Point(v))
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, "}")
	case geom.MultiLineString:
		fmt.Fprintf(w, "{")
		for _, v := range g {
			writeGoGeom(w, geom.LineString(v))
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, "}")
	case geom.Polygon:
		fmt.Fprintf(w, "{")
		for _, v := range g {
			writeGoGeom(w, geom.LineString(v))
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, "}")
	case geom.MultiPolygon:
		fmt.Fprintf(w, "{")
		for _, v := range g {
			writeGoGeom(w, geom.Polygon(v))
			fmt.Fprintf(w, ",")
		}
		fmt.Fprintf(w, "}")
	default:
		return fmt.Errorf("unknown type %T", geo)
	}
	return nil
}

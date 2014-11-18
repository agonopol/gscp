package main

import (
	"flag"
	"gscp/gscplib"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
)

func help() {
	println(`usage: gscp project@bucket:object file
	gscp file project@bucket:object
	gscp project@bucket
	gscp project`)
	os.Exit(1)
}

func main() {
	client := os.Getenv("GSCP_CLIENT_ID")
	if client == "" {
		log.Fatal("Please set GSCP_CLIENT_ID")
	}

	secret := os.Getenv("GSCP_CLIENT_SECRET")
	if secret == "" {
		log.Fatal("Please set GSCP_CLIENT_SECRET")
	}

	cache := os.Getenv("GSCP_CACHE")

	if cache == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		cache = path.Join(u.HomeDir, ".gscp", "cache.json")
		err = os.MkdirAll(path.Dir(cache), 0755)

		if !os.IsExist(err) && err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(cache); err != nil {
			ioutil.WriteFile(cache, []byte{}, 0755)
		}
	}

	gscp := gscplib.NewStore(client, secret, cache, os.Getenv("GSCP_CODE"))

	flag.Parse()
	args := flag.Args()

	if len(args) == 2 {
		if strings.Contains(args[0], "@") {
			gscp.Get(args[0], args[1])
		} else {
			gscp.Put(args[0], args[1])
		}
	} else if len(args) == 1 {
		if strings.Contains(args[0], "@") {
			objects, err := gscp.Ls(args[0])
			if err != nil {
				log.Fatal(err)
			}
			for _, object := range objects {
				println(object.Name)
			}
		} else {
			buckets, err := gscp.Buckets(args[0])
			if err != nil {
				log.Fatal(err)
			}
			for _, bucket := range buckets {
				println(bucket.Id)
			}
		}
	} else {
		help()
	}
}

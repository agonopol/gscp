package main

import (
	"flags"
	"gscp/gscplib"
	"log"
	"os"
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

	secret = os.Getenv("GSCP_CLIENT_SECRET")
	if secret == "" {
		log.Fatal("Please set GSCP_CLIENT_SECRET")
	}

	cache = os.Getenv("GSCP_CACHE")

	gscp := gscplib.NewStore(client, secret, cache)

	args := flags.Args()

	log.Printf("args: %v\n", args)
	if len(args) == 2 {
		if strings.Contains(args[1], "@") {
			gscp.Get(args[1], args[2])
		} else {
			gscp.Put(args[1], args[2])
		}
	} else if len(args) == 1 {
		if strings.Contains(args[1], "@") {
			objects, err := gscp.Ls(args[1])
			if err != nil {
				log.Fatal(err)
			}
			for _, object := range objects {
				println(object)
			}
		} else {
			buckets, err := gscp.Buckets(args[1])
			if err != nil {
				log.Fatal(err)
			}
			for _, bucket := range buckets {
				println(bucket)
			}
		}
	} else {
		help()
	}
}

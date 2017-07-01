package main

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/adhocteam/soapbox/api"
	"github.com/adhocteam/soapbox/api/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 9090, "port to listen on")

	flag.Parse()

	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatalf("couldn't connect to database: %v", err)
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("couldn't listen on port %d: %v", *port, err)
	}

	gs := grpc.NewServer()
	as := api.NewServer(db, nil)
	proto.RegisterApplicationServer(gs, as)
	log.Fatal(gs.Serve(ln))
}

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"

	"github.com/adhocteam/soapbox/api"
	pb "github.com/adhocteam/soapbox/soapboxpb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 9090, "port to listen on")

	flag.Parse()

	if err := checkJobDependencies(); err != nil {
		log.Fatalf("checking for dependencies: %v", err)
	}

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
	pb.RegisterApplicationsServer(gs, as)
	log.Fatal(gs.Serve(ln))
}

func checkJobDependencies() error {
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found: %v", err)
	}
	return nil
}

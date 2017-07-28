package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"

	"github.com/adhocteam/soapbox/api"
	"github.com/adhocteam/soapbox/soapbox"
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

	server := grpc.NewServer()
	config := getConfig()
	if config.AmiId == "" {
		log.Fatal("SOAPBOX_AMI_ID must be set")
	}
	apiServer := api.NewServer(db, nil, config)
	pb.RegisterApplicationsServer(server, apiServer)
	pb.RegisterEnvironmentsServer(server, apiServer)
	pb.RegisterDeploymentsServer(server, apiServer)
	log.Printf("soapboxd listening on 0.0.0.0:%d", *port)
	log.Fatal(server.Serve(ln))
}

func checkJobDependencies() error {
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found: %v", err)
	}
	return nil
}

func getConfig() *soapbox.Config {
	c := &soapbox.Config{
		AmiId:        "", // must be set in the environment
		Domain:       "soapbox.hosting",
		IamProfile:   "soapbox-app",
		InstanceType: "t2.micro",
		KeyName:      "soapbox-app",
	}
	if val := os.Getenv("SOAPBOX_AMI_ID"); val != "" {
		c.AmiId = val
	}
	if val := os.Getenv("SOAPBOX_DOMAIN"); val != "" {
		c.Domain = val
	}
	if val := os.Getenv("SOAPBOX_IAM_PROFILE"); val != "" {
		c.IamProfile = val
	}
	if val := os.Getenv("SOAPBOX_INSTANCE_TYPE"); val != "" {
		c.InstanceType = val
	}
	if val := os.Getenv("SOAPBOX_KEY_NAME"); val != "" {
		c.KeyName = val
	}
	return c
}

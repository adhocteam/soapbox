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
	"time"

	"github.com/adhocteam/soapbox"
	"github.com/adhocteam/soapbox/buildinfo"
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 9090, "port to listen on")
	printVersion := flag.Bool("V", false, "print version and exit")

	flag.Parse()

	if *printVersion {
		fmt.Printf("        version: %s\n", buildinfo.Version)
		fmt.Printf("     git commit: %s\n", buildinfo.GitCommit)
		fmt.Printf("     build time: %s\n", buildinfo.BuildTime)
		return
	}

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

	server := grpc.NewServer(serverInterceptor())
	config := getConfig()
	apiServer := soapboxd.NewServer(db, nil, config)
	pb.RegisterApplicationsServer(server, apiServer)
	pb.RegisterConfigurationsServer(server, apiServer)
	pb.RegisterEnvironmentsServer(server, apiServer)
	pb.RegisterDeploymentsServer(server, apiServer)
	pb.RegisterUsersServer(server, apiServer)
	pb.RegisterVersionServer(server, apiServer)
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
		AmiName:      "soapbox-aws-linux-ami-*",
		Domain:       "soapbox.hosting",
		IamProfile:   "soapbox-app",
		InstanceType: "t2.micro",
		KeyName:      "soapbox-app",
	}
	if val := os.Getenv("SOAPBOX_AMI_NAME"); val != "" {
		c.AmiName = val
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

func serverInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(grpc.UnaryServerInterceptor(timingInterceptor))
}

func timingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	t0 := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("method=%s duration=%s error=%v", info.FullMethod, time.Since(t0), err)
	return resp, err
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adhocteam/soapbox/api/proto"

	"google.golang.org/grpc"
)

func main() {
	serverAddr := flag.String("server", "127.0.0.1:9090", "host:port of server")

	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("couldn't connect to server %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	client := proto.NewApplicationClient(conn)
	ctx := context.Background()

	type command func(context.Context, proto.ApplicationClient, []string) error
	var cmd command

	switch flag.Arg(0) {
	case "create-application":
		cmd = createApplication
	default:
		log.Fatalf("unknown command %q", flag.Arg(0))
	}

	if err := cmd(ctx, client, flag.Args()); err != nil {
		log.Fatalf("error executing command %s: %v", flag.Arg(0), err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <command>\n", filepath.Base(os.Args[0]))
}

func createApplication(ctx context.Context, client proto.ApplicationClient, args []string) error {
	args = args[1:]
	if len(args) < 4 {
		return fmt.Errorf("4 arguments are required: name, description, github repo URL, and type (server, cronjob)")
	}

	var appType proto.CreateApplicationRequest_Type
	switch args[3] {
	case "server":
		appType = proto.CreateApplicationRequest_SERVER
	case "cronjob":
		appType = proto.CreateApplicationRequest_CRONJOB
	default:
		return fmt.Errorf("unknown app type %q, expecting either 'server' or 'cronjob'", args[3])
	}

	req := &proto.CreateApplicationRequest{
		Name:          args[0],
		Description:   args[1],
		GithubRepoURL: args[2],
		Type:          appType,
	}
	app, err := client.CreateApplication(ctx, req)
	if err != nil {
		return fmt.Errorf("error creating application: %v", err)
	}
	log.Printf("created application %q, ID %d", args[0], app.GetId())
	return nil
}

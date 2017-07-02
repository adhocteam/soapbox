package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	pb "github.com/adhocteam/soapbox/soapboxpb"

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

	client := pb.NewApplicationClient(conn)
	ctx := context.Background()

	type command func(context.Context, pb.ApplicationClient, []string, io.Writer) error
	var cmd command

	switch flag.Arg(0) {
	case "create-application":
		cmd = createApplication
	case "list-applications":
		cmd = listApplications
	default:
		log.Fatalf("unknown command %q", flag.Arg(0))
	}

	if err := cmd(ctx, client, flag.Args(), os.Stdout); err != nil {
		log.Fatalf("error executing command %s: %v", flag.Arg(0), err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <command>\n", filepath.Base(os.Args[0]))
}

func createApplication(ctx context.Context, client pb.ApplicationClient, args []string, stdout io.Writer) error {
	args = args[1:]
	if len(args) < 4 {
		return fmt.Errorf("4 arguments are required: name, description, github repo URL, and type (server, cronjob)")
	}

	var appType pb.CreateApplicationRequest_Type
	switch args[3] {
	case "server":
		appType = pb.CreateApplicationRequest_SERVER
	case "cronjob":
		appType = pb.CreateApplicationRequest_CRONJOB
	default:
		return fmt.Errorf("unknown app type %q, expecting either 'server' or 'cronjob'", args[3])
	}

	req := &pb.CreateApplicationRequest{
		Name:          args[0],
		Description:   args[1],
		GithubRepoURL: args[2],
		Type:          appType,
	}
	app, err := client.CreateApplication(ctx, req)
	if err != nil {
		return fmt.Errorf("error creating application: %v", err)
	}

	fmt.Fprintf(stdout, "created application %q, ID %d", args[0], app.GetId())
	return nil
}

func listApplications(ctx context.Context, client pb.ApplicationClient, args []string, stdout io.Writer) error {
	req := &pb.ListApplicationRequest{}
	apps, err := client.ListApplications(ctx, req)
	if err != nil {
		return fmt.Errorf("error listing applications: %v", err)
	}
	for _, app := range apps.Applications {
		fmt.Fprintf(stdout, "%d\t%s\n", app.Id, app.Name)
	}
	return nil
}

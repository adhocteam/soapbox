package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

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

	client := pb.NewApplicationsClient(conn)
	ctx := context.Background()

	type command func(context.Context, pb.ApplicationsClient, []string) error
	var cmd command

	switch flag.Arg(0) {
	case "create-application":
		cmd = createApplication
	case "list-applications":
		cmd = listApplications
	case "get-application":
		cmd = getApplication
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

func createApplication(ctx context.Context, client pb.ApplicationsClient, args []string) error {
	args = args[1:]
	if len(args) < 4 {
		return fmt.Errorf("4 arguments are required: name, description, github repo URL, and type (server, cronjob)")
	}

	var appType pb.ApplicationType
	switch args[3] {
	case "server":
		appType = pb.ApplicationType_SERVER
	case "cronjob":
		appType = pb.ApplicationType_CRONJOB
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

	fmt.Printf("created application %q, ID %d", args[0], app.GetId())
	return nil
}

func listApplications(ctx context.Context, client pb.ApplicationsClient, args []string) error {
	req := &pb.ListApplicationRequest{}
	apps, err := client.ListApplications(ctx, req)
	if err != nil {
		return fmt.Errorf("error listing applications: %v", err)
	}
	for _, app := range apps.Applications {
		fmt.Printf("%d\t%s\n", app.Id, app.Name)
	}
	return nil
}

func getApplication(ctx context.Context, client pb.ApplicationsClient, args []string) error {
	args = args[1:]
	if len(args) < 1 {
		return fmt.Errorf("1 argument required: ID of application")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %v", err)
	}
	req := &pb.GetApplicationRequest{Id: int32(id)}
	resp, err := client.GetApplication(ctx, req)
	if err != nil {
		return fmt.Errorf("getting application: %v", err)
	}
	fmt.Printf("name:                %s\n", resp.App.Name)
	fmt.Printf("ID:                  %d\n", resp.App.Id)
	fmt.Printf("type:                %s\n", pb.ApplicationType_name[int32(resp.App.Type)])
	fmt.Printf("created at:          %s\n", resp.App.CreatedAt)
	fmt.Printf("external DNS:        %s\n", resp.App.ExternalDNS)
	fmt.Printf("Dockerfile path:     %s\n", resp.App.DockerfilePath)
	fmt.Printf("entrypoint override: %s\n", resp.App.EntrypointOverride)
	fmt.Printf("description:\n%s\n", resp.App.Description)
	return nil
}

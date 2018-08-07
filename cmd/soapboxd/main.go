package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/adhocteam/soapbox"
	"github.com/adhocteam/soapbox/buildinfo"
	pb "github.com/adhocteam/soapbox/proto"
	"github.com/adhocteam/soapbox/soapboxd"
	"github.com/adhocteam/soapbox/soapboxd/aws"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	port := flag.Int("port", 9090, "port to listen on")
	printVersion := flag.Bool("V", false, "print version and exit")
	logTiming := flag.Bool("log-timing", false, "print log of method call timings")

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

	var opts []grpc.ServerOption
	opts = append(opts, serverInterceptor(loginInterceptor))
	if *logTiming {
		opts = append(opts, serverInterceptor(timingInterceptor))
	}

	soapboxConfig := getConfig()
	cloud := aws.NewCloudProvider(soapboxConfig)

	server := grpc.NewServer(opts...)
	apiServer := soapboxd.NewServer(db, nil, cloud)
	pb.RegisterApplicationsServer(server, apiServer)
	pb.RegisterConfigurationsServer(server, apiServer)
	pb.RegisterEnvironmentsServer(server, apiServer)
	pb.RegisterDeploymentsServer(server, apiServer)
	pb.RegisterUsersServer(server, apiServer)
	pb.RegisterVersionServer(server, apiServer)
	pb.RegisterActivitiesServer(server, apiServer)
	log.Printf("soapboxd listening on 0.0.0.0:%d", *port)
	log.Fatal(server.Serve(ln))
}

func checkJobDependencies() error {
	binaries := []string{"terraform", "docker", "git"}
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("%s not found: %v", bin, err)
		}
	}
	return nil
}

func getConfig() soapbox.Config {
	c := soapbox.Config{
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

func serverInterceptor(interceptor grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(interceptor)
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

func loginInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	switch strings.Split(info.FullMethod, "/")[2] {
	case "LoginUser", "CreateUser", "GetUser":
		return handler(ctx, req)
	default:
		if err := authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

type accessDeniedErr struct {
	userID []byte
}

func (e *accessDeniedErr) Error() string {
	return fmt.Sprintf("Incorrect login token for user %s", e.userID)
}

type missingMetadataErr struct{}

func (e *missingMetadataErr) Error() string {
	return fmt.Sprint("Not enough metadata attached to authorize request")
}

// TODO(kalilsn) The token calculated here is static, so it can't be revoked, and if stolen
// would allow an attacker to impersonate a user indefinitely.
func authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals, ok := md["user_id"]
		if !ok || len(vals) < 1 {
			return &missingMetadataErr{}
		}
		userID := []byte(vals[0])
		vals, ok = md["login_token"]
		if !ok || len(vals) < 1 {
			return &missingMetadataErr{}
		}
		sentToken, err := base64.StdEncoding.DecodeString(vals[0])
		if err != nil {
			return err
		}
		key := []byte(os.Getenv("LOGIN_SECRET_KEY"))
		h := hmac.New(sha512.New, key)
		h.Write(userID)
		calculated := h.Sum(nil)
		if hmac.Equal(sentToken, calculated) {
			return nil
		}

		return &accessDeniedErr{userID}
	}

	return &missingMetadataErr{}
}

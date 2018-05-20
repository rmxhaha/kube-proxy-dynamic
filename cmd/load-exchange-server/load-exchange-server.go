package main

import (
	"flag"
	"time"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"log"

	"net"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/exchange/server"
	pb "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/exchange/loadexchange"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert-file", "", "The TLS cert file")
	keyFile    = flag.String("key-file", "", "The TLS key file")
	port       = flag.Int("port", 14156, "The server port")
	kubeconfig = flag.String("kubeconfig","/var/lib/load-exchange-server/kubeconfig", "Kubeconfig to access kubernetes API")
	updateInterval = flag.Duration("update-interval", 500 * time.Millisecond,"how often should server sends podloads")
)

func main(){
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	loadExchangeServer, err := server.New(clientset, *updateInterval)
	if err != nil {
		log.Fatal(err)
	}

	go loadExchangeServer.Run()


	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	if *tls {
		if *certFile == "" {
			log.Fatalf("certFile not provided")
		}
		if *keyFile == "" {
			log.Fatalf("keyFile not provided")
		}

		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterLoadExchangeServer(grpcServer, loadExchangeServer)
	grpcServer.Serve(lis)

}

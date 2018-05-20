package main

import (
	"flag"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"time"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/loadbalancer"
)

var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca-file", "", "The file containning the CA root cert file")
	serverHostOverride = flag.String("server-host-override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
	port       = flag.Int("port", 14156, "the server port")
	kubeconfig = flag.String("kubeconfig","/var/lib/load-exchange-server/kubeconfig", "Kubeconfig to access kubernetes API")
	updateInterval = flag.Duration("update-interval",500 * time.Millisecond, "IPVS Synchronize period")
)


func main(){

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			log.Fatalf("caFile not provided")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	lb, err := loadbalancer.NewLoadBalancer(clientset, *port, opts)
	if err != nil {
		log.Fatal(err)
	}

	lb.SyncLoop(*updateInterval)
}
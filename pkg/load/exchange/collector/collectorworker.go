package collector

import (
	"google.golang.org/grpc"
	"io"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"log"
	"time"
	pb "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/exchange/loadexchange"
	"net"
	"context"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
)


type worker struct {
	podLoadStore *store.Store
	targetURL string
	ShouldQuit bool
}

func newWorker(TargetURL string, podLoadStore *store.Store) (*worker, error) {
	return &worker{ ShouldQuit: false, targetURL: TargetURL, podLoadStore: podLoadStore}, nil
}

func (lcw *worker) CollectTillDie() error {
	var opts []grpc.DialOption
	opts = append( opts, grpc.WithInsecure())
	conn, err := grpc.Dial(lcw.targetURL, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewLoadExchangeClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	stream, err := client.GetPodLoads(ctx, &pb.PodSelector{})

	if err != nil {
		return  err
	}

	waitc := make(chan struct{})
	go func(){
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			if err != nil {
				fmt.Printf("Failed to receieve with a note: %v", err)
				close(waitc)
				break
			}


			fmt.Printf("[%s] New Info\n", lcw.targetURL)
			lcw.ApplyToStore(in, lcw.targetURL)
		}
	}()

	stream.CloseSend()

	<- waitc
	return nil
}

func (lcw *worker) ApplyToStore(podLoads *pb.PodLoads,  HostIP string){
	rt, err := ptypes.Timestamp(podLoads.RecordTime)
	if err != nil {
		log.Print(err)
		return
	}
	for _, pl := range podLoads.PodLoads {

		if len(pl.PodIP) != 4 {
			log.Printf("Wrong PodIP format (%d): %s", len(pl.PodIP), pl.PodIP)
			continue
		}

		ip_str := net.IPv4(pl.PodIP[0],pl.PodIP[1],pl.PodIP[2],pl.PodIP[3]).String()


		lcw.podLoadStore.Add(&store.PodLoad{ PodIP: ip_str, Load: pl.Load, RecordTime: rt, HostIP:HostIP })
	}
}

func (lcw *worker) Run(){
	for {
		fmt.Printf("[%s] CollectTillDie", lcw.targetURL)
		err := lcw.CollectTillDie()
		if err != nil {
			log.Println(err)
		}

		if lcw.ShouldQuit {
			break
		}

		time.Sleep(1 * time.Second)
	}
}

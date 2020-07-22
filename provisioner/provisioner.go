// Copyright 2018 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provisioner

import (
	"context"
	"fmt"
	//"golang.org/x/net/context"

	//"github.com/container-storage-interface/spec/lib/go/csi"
	csi "github.com/sodafoundation/etcd-watch-go-client/csimodel"
	"time"

	log "github.com/golang/glog"
	//"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
	//context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"os/exec"
)

// Client interface provides an abstract description about how to interact
// with gRPC client. Besides some nested methods defined in pb.ControllerClient,
// Client also exposes two methods: Connect() and Close(), for which callers
// can easily open and close gRPC connection.
type Client1 interface {
	csi.ControllerClient
	Connect(edp string) error
}

// client structure is one implementation of Client interface and will be
// called in real environment. There would be more other kind of connection
// in the long run.
type client1 struct {
	csi.ControllerClient
	*grpc.ClientConn
}

func NewClient() Client1 { return &client1{} }

func (c *client1) Connect(edp string) error {
	fmt.Print("edp===:", edp)
	// Set up a connection to the Dock server.
	if c.ClientConn != nil && c.ClientConn.GetState() == connectivity.Ready {
		return nil
	}
	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
	conn, err := grpc.Dial(edp, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Errorf("did not connect: %+v\n", err)
		return err
	}
	// Create controller client via the connection.
	c.ControllerClient = csi.NewControllerClient(conn)
	c.ClientConn = conn

	return nil
}

func NewVolumePortal() *VolumePortal {
	return &VolumePortal{
		CtrClient: NewClient(),
	}
}

type VolumePortal struct {
	CtrClient Client1
}

func ProcCsiRequest(req *csi.CreateVolumeRequest, csiEndpoint *string) (*csi.CreateVolumeResponse, error) {

	//out, err := exec.Command("ls -l /var/lib/kubelet/plugins/csi-lvmplugin/").Output()
	out, err := exec.Command("ls", "-ltr").Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Println("Command Successfully Executed")
	output := string(out[:])
	fmt.Println(output)

	var edp = "unix:///var/lib/kubelet/plugins/csi-lvmplugin/csi.sock"
	var v = VolumePortal{CtrClient: NewClient()}
	if err := v.CtrClient.Connect(edp); err != nil {
		log.Error("when connecting controller client:", err)
		return nil, err
	}
	//opt := &csi.CreateVolumeRequest{
	//	Name:             "Dummy",
	//	// TODO: ProfileId will be removed later.
	//}
	//response, err := v.CtrClient.CreateVolume(context.Background(), opt)
	response, err := v.CtrClient.CreateVolume(context.Background(), req)
	if err != nil {
		log.Error("create volume failed in controller service:", err)
		return nil, err
	}
	log.Info(response)
	fmt.Print("response...===:", response)
	return response, nil
}

//import (
//	"context"
//	"fmt"
//	"github.com/container-storage-interface/spec/lib/go/csi"
//	"github.com/golang/glog"
//	"google.golang.org/grpc/connectivity"
//	//pb "github.com/sodafoundation/api/pkg/model/proto"
//	"net"
//	"strings"
//
//	//"github.com/kubernetes-csi/csi-lib-utils/connection"
//	//"os"
//
//	//"os"
//
//	"google.golang.org/grpc"
//	//"google.golang.org/grpc/keepalive"
//	"github.com/kubernetes-csi/csi-lib-utils/metrics"
//	//"github.com/kubernetes-csi/csi-lib-utils/rpc"
//	//"os"
//	"time"
//)
//
//func logGRPC(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
//	glog.V(5).Infof("GRPC call: %s", method)
//	glog.V(5).Infof("GRPC request: %+v", req)
//	err := invoker(ctx, method, req, reply, cc, opts...)
//	glog.V(5).Infof("GRPC response: %+v", reply)
//	glog.V(5).Infof("GRPC error: %v", err)
//	return err
//}
//
//func connect(address string, timeout time.Duration) (*grpc.ClientConn, error) {
//	glog.V(2).Infof("Connecting to %s", address)
//	dialOptions := []grpc.DialOption{
//		grpc.WithInsecure(),
//		grpc.WithBackoffMaxDelay(time.Second),
//		grpc.WithUnaryInterceptor(logGRPC),
//	}
//	if strings.HasPrefix(address, "/") {
//		dialOptions = append(dialOptions, grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
//			return net.DialTimeout("unix", addr, timeout)
//		}))
//	}
//	conn, err := grpc.Dial(address, dialOptions...)
//    fmt.Println("conn..........:",conn)
//	if err != nil {
//		return nil, err
//	}
//	ctx, cancel := context.WithTimeout(context.Background(), timeout)
//	defer cancel()
//	for {
//		if !conn.WaitForStateChange(ctx, conn.GetState()) {
//			glog.V(4).Infof("Connection timed out")
//			return conn, nil // return nil, subsequent GetPluginInfo will show the real connection error
//		}
//		if conn.GetState() == connectivity.Ready {
//			glog.V(3).Infof("Connected")
//			return conn, nil
//		}
//		glog.V(4).Infof("Still trying, connection is %s", conn.GetState())
//	}
//}
//
//type Client interface {
//	csi.ControllerClient
//
//	Connect(edp string) error
//}
//
//
//// NewServer Initialises and return plugin server
//func ProcCsiRequest(req *csi.CreateVolumeRequest, csiEndpoint *string) (error) {
//	// When there are multiple volumes unmount at the same time,
//	// it will cause conflicts related to the state machine,
//	// so start a watch list to let the volumes unmount one by one.
//	fmt.Println("Inside called")
//	metricsManager := metrics.NewCSIMetricsManager("csi-lvmplugin") /* driverName */
//	operationTimeout := 10*time.Second
//
//	//var kacp = keepalive.ClientParameters{
//	//	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
//	//	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
//	//	PermitWithoutStream: true,             // send pings even without active streams
//	//}
//
//	//csiEndpoint1 := "unix://var/lib/kubelet/plugins/csi-opensdsplugin-block/csi.sock"
//	csiEndpoint1:= "/var/lib/kubelet/plugins/csi-lvmplugin/csi.sock"
//    //csiEndpoint1 := "/var/lib/kubelet/plugins/csi-lvmplugin/csi.sock"
//	fmt.Println("csiEndpoint....", csiEndpoint1)
//
//	grpcClient, err := connect(csiEndpoint1, 10)
//	//grpcClient, err := grpc.Dial(csiEndpoint1, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
//	//if err != nil {
//	//	fmt.Println("Dial failed", csiEndpoint)
//	//	return err
//	//}
//
//		//grpcClient, err := Connect(*csiEndpoint, metricsManager)
//		//if err != nil {
//		//	fmt.Println("Connect failed")
//		//	os.Exit(1)
//		//}
//
//	//err = Probe(grpcClient, operationTimeout)
//	//if err != nil {
//	//	fmt.Println("Probe failed")
//	//	os.Exit(1)
//	//}
//	// 0.0.0.0:1736
//	// Autodetect provisioner name
//	provisionerName := "csi-lvmplugin"
//	//provisionerName := "csi-opensdsplugin-block"
//
//	fmt.Println("Detected CSI driver %s", provisionerName)
//	metricsManager.SetDriverName(provisionerName)
//	//metricsManager.StartMetricsEndpoint(*metricsAddress, *metricsPath)
//
//	ctx, _ := context.WithTimeout(context.Background(), operationTimeout)
//	//defer cancel()
//
//	csiClient := csi.NewControllerClient(grpcClient)
//
//	fmt.Println("csiClient rep %v", csiClient)
//
//	//rep, err := csiClient.CreateVolume(ctx, req)
//	rep, err := Client.CreateVolume(context.Background(), req)
//	fmt.Println("CreateVolume rep %v", rep)
//	if err != nil {
//		fmt.Println("CreateVolume failed", err)
//		return err
//	}
//
//	if rep.Volume != nil {
//		fmt.Println("CreateVolume failed2",*rep.Volume)
//	}
//
//	return nil
//}
//
////func Connect(address string, metricsManager metrics.CSIMetricsManager) (*grpc.ClientConn, error) {
////	return connection.Connect(address, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss())), nil
////}
////
////func Probe(conn *grpc.ClientConn, singleCallTimeout time.Duration) error {
////	return rpc.ProbeForever(conn, singleCallTimeout)
////}

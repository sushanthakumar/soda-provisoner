package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/golang/glog"

	//"github.com/container-storage-interface/spec/lib/go/csi"
	csi "github.com/sodafoundation/etcd-watch-go-client/csimodel"
	"github.com/sodafoundation/etcd-watch-go-client/provisioner"
	"go.etcd.io/etcd/v3/clientv3"
	"time"
)
type BaseModel struct {
	// The uuid of the object, it's unique in the context and generated by system
	// on successful creation of the object. It's not allowed to be modified by
	// the user.
	// +readOnly
	Id string `json:"id"`

	// CreateAt representing the server time when the object was created successfully.
	// Now, it's represented as a time string in RFC8601 format.
	// +readOnly
	CreatedAt string `json:"createdAt"`

	// UpdatedAt representing the server time when the object was updated successfully.
	// Now, it's represented as a time string in RFC8601 format.
	// +readOnly
	UpdatedAt string `json:"updatedAt"`
}

type VolumeContentSource_VolumeSource struct {
	// Contains identity information for the existing source volume.
	// This field is REQUIRED. Plugins reporting CLONE_VOLUME
	// capability MUST support creating a volume from another volume.
	VolumeId             string   `protobuf:"bytes,1,opt,name=volume_id,json=volumeId,proto3" json:"volume_id,omitempty"`
}
type VolumeContentSource_Volume struct {
	Volume *VolumeContentSource_VolumeSource `protobuf:"bytes,2,opt,name=volume,proto3,oneof"`
}

type isVolumeContentSource_Type *VolumeContentSource_Volume

type VolumeContentSource struct {
	// Types that are valid to be assigned to Type:
	//	*VolumeContentSource_Snapshot
	//	*VolumeContentSource_Volume
	Type isVolumeContentSource_Type `protobuf_oneof:"type"`
}

type VolumeCapability struct {
	// Specifies what API the volume will be accessed using. One of the
	// following fields MUST be specified.
	//
	// Types that are valid to be assigned to AccessType:
	//	*VolumeCapability_Block
	//	*VolumeCapability_Mount
	//AccessType isVolumeCapability_AccessType `protobuf_oneof:"access_type"`
	// This is a REQUIRED field.
	AccessMode  *VolumeCapability_AccessMode `protobuf:"bytes,3,opt,name=access_mode,json=accessMode,proto3" json:"access_mode,omitempty"`
}

type isVolumeCapability_AccessType interface {
	isVolumeCapability_AccessType()
}
type VolumeCapability_AccessMode_Mode int32

type VolumeCapability_AccessMode struct {
	// This field is REQUIRED.
	Mode VolumeCapability_AccessMode_Mode `protobuf:"varint,1,opt,name=mode,proto3,enum=csi.v1.VolumeCapability_AccessMode_Mode" json:"mode,omitempty"`
}

type CapacityRange struct {
	// Volume MUST be at least this big. This field is OPTIONAL.
	// A value of 0 is equal to an unspecified field value.
	// The value of this field MUST NOT be negative.
	RequiredBytes int64 `protobuf:"varint,1,opt,name=required_bytes,json=requiredBytes,proto3" json:"required_bytes,omitempty"`
	// Volume MUST not be bigger than this. This field is OPTIONAL.
	// A value of 0 is equal to an unspecified field value.
	// The value of this field MUST NOT be negative.
	LimitBytes           int64    `protobuf:"varint,2,opt,name=limit_bytes,json=limitBytes,proto3" json:"limit_bytes,omitempty"`
}

type CsiVolumeSpec struct {
	*BaseModel
	//// The name of the volume.
	//Name string `json:"name,omitempty"`
	//// Default unit of volume Size is GB.
	//Size int64 `json:"size,omitempty"`
	//*csi.CreateVolumeRequest
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`

	CapacityRange *CapacityRange `protobuf:"bytes,2,opt,name=capacity_range,json=capacityRange,proto3" json:"capacity_range,omitempty"`
	// This field is REQUIRED.
	VolumeCapabilities []*VolumeCapability `protobuf:"bytes,3,rep,name=volume_capabilities,json=volumeCapabilities,proto3" json:"volume_capabilities,omitempty"`
	// Plugin specific parameters passed in as opaque key-value pairs.
	// This field is OPTIONAL. The Plugin is responsible for parsing and
	// validating these parameters. COs will treat these as opaque.
	Parameters map[string]string `protobuf:"bytes,4,rep,name=parameters,proto3" json:"parameters,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Secrets required by plugin to complete volume creation request.
	// This field is OPTIONAL. Refer to the `Secrets Requirements`
	// section on how to use this field.
	Secrets map[string]string `protobuf:"bytes,5,rep,name=secrets,proto3" json:"secrets,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// If specified, the new volume will be pre-populated with data from
	// this source. This field is OPTIONAL.
	VolumeContentSource *VolumeContentSource `protobuf:"bytes,6,opt,name=volume_content_source,json=volumeContentSource,proto3" json:"volume_content_source,omitempty"`
}

func main() {
	etcdHost := flag.String("etcdHost", "10.0.2.15:62379", "etcd host")
	etcdWatchKey := flag.String("etcdWatchKey", "/opensds/v1beta/block/volumes/e93b4c0934da416eb9c8d120c5d04d96/4a33e3fb-d3de-4faa-a192-8b29c0f67846", "etcd key to watch")
	csiEndpoint   := flag.String("csi-address", "/var/lib/kubelet/plugins/csi-lvmplugin/csi.sock", "The gRPC endpoint for Target CSI Volume.")
	//csiEndpoint   := flag.String("csi-address", "abcd", "The gRPC endpoint for Target CSI Volume.")

	flag.Parse()

	fmt.Println("connecting to etcd.. - " + *etcdHost)

	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://" + *etcdHost},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("connected to etcd - " + *etcdHost)

	defer etcd.Close()

	watchChan := etcd.Watch(context.Background(), *etcdWatchKey)
	fmt.Println("set WATCH on " + *etcdWatchKey)

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			fmt.Printf("Event received! %s executed on %q with value %q\n", event.Type, event.Kv.Key, event.Kv.Value)
			fmt.Println("ProcCsiRequest called with  event.Kv.Key",  event.Kv.Key)

			//byt := []byte(event.Kv.Value}

			var obj csi.CreateVolumeRequest
			if err := json.Unmarshal(event.Kv.Value, &obj); err != nil {
				fmt.Println("ProcCsiRequest Unmarshal failed")
				panic(err)
			}
			fmt.Println("ProcCsiRequest called with  name",  obj.Name)
			//fmt.Println("ProcCsiRequest called with  RequiredBytes",  obj.CapacityRange.RequiredBytes)

/*			volumeReq := csi.CreateVolumeRequest{
				Name:obj.Name,
				CapacityRange:obj.CapacityRange,
				VolumeCapabilities: obj.VolumeCapabilities,
				VolumeContentSource: obj.VolumeContentSource,
				Parameters: obj.Parameters,
				//Secrets: obj.Secrets,
			}*/
			fmt.Println("ProcCsiRequest called with  name",  obj.Name)
			response, err := provisioner.ProcCsiRequest(&obj, csiEndpoint)
			log.Info(response)
			log.Info(err)
		}
	}

}

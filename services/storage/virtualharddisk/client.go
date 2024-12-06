// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the Apache v2.0 License.

package virtualharddisk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/microsoft/moc-sdk-for-go/services/compute"
	"github.com/microsoft/moc-sdk-for-go/services/storage"
	"github.com/microsoft/moc/pkg/auth"
	"github.com/microsoft/moc/pkg/errors"
	"github.com/microsoft/moc/rpc/common"
)

// Service interface
type Service interface {
	Get(context.Context, string, string, string) (*[]storage.VirtualHardDisk, error)
	CreateOrUpdate(context.Context, string, string, string, *storage.VirtualHardDisk, string, common.ImageSource) (*storage.VirtualHardDisk, error)
	Delete(context.Context, string, string, string) error
	Precheck(context.Context, string, string, []*storage.VirtualHardDisk) (bool, error)
	Upload(context.Context, string, string, *storage.VirtualHardDisk, string) error
}

// Client structure
type VirtualHardDiskClient struct {
	storage.BaseClient
	internal Service
}

// NewClient method returns new client
func NewVirtualHardDiskClient(cloudFQDN string, authorizer auth.Authorizer) (*VirtualHardDiskClient, error) {
	c, err := newVirtualHardDiskClient(cloudFQDN, authorizer)
	if err != nil {
		return nil, err
	}

	return &VirtualHardDiskClient{internal: c}, nil
}

// Get methods invokes the client Get method
func (c *VirtualHardDiskClient) Get(ctx context.Context, group, container, name string) (*[]storage.VirtualHardDisk, error) {
	return c.internal.Get(ctx, group, container, name)
}

// CreateOrUpdate methods invokes create or update on the client
func (c *VirtualHardDiskClient) CreateOrUpdate(ctx context.Context, group, container, name string, storage *storage.VirtualHardDisk) (*storage.VirtualHardDisk, error) {
	return c.internal.CreateOrUpdate(ctx, group, container, name, storage, "", common.ImageSource_LOCAL_SOURCE)
}

// Delete methods invokes delete of the storage resource
func (c *VirtualHardDiskClient) Delete(ctx context.Context, group, container, name string) error {
	return c.internal.Delete(ctx, group, container, name)
}

// Resize methods invokes Update to change size of the storage resource
func (c *VirtualHardDiskClient) Resize(ctx context.Context, group, container, name string, newSize int64) error {
	vhds, err := c.Get(ctx, group, container, name)
	if err != nil {
		return err
	}

	if len(*vhds) == 0 {
		return errors.Wrapf(errors.NotFound, "%s", name)
	}

	vhd := (*vhds)[0]
	vhd.DiskSizeBytes = &newSize

	_, err = c.CreateOrUpdate(ctx, group, container, name, &vhd)

	return err
}

// Upload methods invokes upload of the storage resource to target sasurl
func (c *VirtualHardDiskClient) Upload(ctx context.Context, group, container, name string, targetUrl string) error {
	fmt.Printf("moc-sdk-for-go: client.go: Upload start: vhd name [%s] to [%s]\n", name, targetUrl)
	fmt.Printf("moc-sdk-for-go: client.go: Calling Get on VHD name\n")

	vhds, err := c.Get(ctx, group, container, name)
	if err != nil {
		return err
	}
	fmt.Printf("moc-sdk-for-go: client.go: Done with Get\n")

	if vhds == nil || len(*vhds) == 0 {
		return errors.Wrapf(errors.NotFound, "%s", name)
	}
	fmt.Printf("moc-sdk-for-go: client.go: Found vhd as len is not zero\n")

	if targetUrl == "" {
		return errors.Wrapf(errors.InvalidInput, "targetUrl cannot be empty")
	}
	fmt.Printf("moc-sdk-for-go: client.go: Targeturl is not empty\n")

	vhd := (*vhds)[0]
	fmt.Printf("moc-sdk-for-go: client.go: Calling internal Upload now\n")

	return c.internal.Upload(ctx, group, container, &vhd, targetUrl)
}

// Prechecks whether the system is able to create specified virtual hard disks.
// Returns true with virtual hard disk placement in mapping from virtual hard disk names to container names; or false with reason in error message.
func (c *VirtualHardDiskClient) Precheck(ctx context.Context, group, container string, vhds []*storage.VirtualHardDisk) (bool, error) {
	return c.internal.Precheck(ctx, group, container, vhds)
}

func (c *VirtualHardDiskClient) DownloadVhdFromHttp(ctx context.Context, group, container, name string, storage *storage.VirtualHardDisk, azHttpImg *compute.AzureGalleryImageProperties) (*storage.VirtualHardDisk, error) {
	// convert httpImg struct to json string and use it as image-path
	data, err := json.Marshal(azHttpImg)
	if err != nil {
		return nil, err
	}
	datastring := string(data)
	return c.internal.CreateOrUpdate(ctx, group, container, name, storage, datastring, common.ImageSource_HTTP_SOURCE)
}

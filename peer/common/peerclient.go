/*
Copyright IBM Corp. 2016-2017 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package common

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/comm"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// PeerClient represents a client for communicating with a peer
type PeerClient struct {
	commonClient
}

// NewPeerClientFromEnv creates an instance of a PeerClient from the global
// Viper instance
//创建peer客户端
func NewPeerClientFromEnv() (*PeerClient, error) {
	//获取环境变量
	address, override, clientConfig, err := configFromEnv("peer")
	if err != nil {
		return nil, errors.WithMessage(err,
			"failed to load config for PeerClient")
	}
	// 设置超市3秒
	clientConfig.Timeout = time.Second * 3
	//创建grpc客户端
	gClient, err := comm.NewGRPCClient(clientConfig)
	if err != nil {
		return nil, errors.WithMessage(err,
			"failed to create PeerClient from config")
	}
	//创建peer客户端
	pClient := &PeerClient{
		commonClient: commonClient{
			GRPCClient: gClient,
			address:    address,
			sn:         override}}
	return pClient, nil
}

// Endorser returns a client for the Endorser service
// 获取背书节点
func (pc *PeerClient) Endorser() (pb.EndorserClient, error) {
	//获取grpc链接
	conn, err := pc.commonClient.NewConnection(pc.address, pc.sn)
	if err != nil {
		return nil, errors.WithMessage(err,
			fmt.Sprintf("endorser client failed to connect to %s", pc.address))
	}
	// 创建背书客户端
	return pb.NewEndorserClient(conn), nil
}

// Admin returns a client for the Admin service
func (pc *PeerClient) Admin() (pb.AdminClient, error) {
	conn, err := pc.commonClient.NewConnection(pc.address, pc.sn)
	if err != nil {
		return nil, errors.WithMessage(err,
			fmt.Sprintf("admin client failed to connect to %s", pc.address))
	}
	return pb.NewAdminClient(conn), nil
}

// GetEndorserClient returns a new endorser client.  The target address for
// the client is taken from the configuration setting "peer.address"
func GetEndorserClient() (pb.EndorserClient, error) {
	// 根据peer环境变量创建peer客户端
	peerClient, err := NewPeerClientFromEnv()
	if err != nil {
		return nil, err
	}
	return peerClient.Endorser()
}

// GetAdminClient returns a new admin client.  The target address for
// the client is taken from the configuration setting "peer.address"
func GetAdminClient() (pb.AdminClient, error) {
	peerClient, err := NewPeerClientFromEnv()
	if err != nil {
		return nil, err
	}
	return peerClient.Admin()
}

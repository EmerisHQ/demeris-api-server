package ibcclient

import (
	"fmt"

	"context"

	ibcchannel "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"google.golang.org/grpc"
)

var grpcport = 9090

func IbcChannelClientState(chainName string, channelId, portId string) (*ibcchannel.QueryChannelClientStateResponse, error) {
	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcport), grpc.WithInsecure())
	if err != nil {
		return &ibcchannel.QueryChannelClientStateResponse{}, err
	}

	defer func() {
		_ = grpcConn.Close()
	}()

	iq := ibcchannel.NewQueryClient(grpcConn)
	return iq.ChannelClientState(context.Background(), &ibcchannel.QueryChannelClientStateRequest{
		ChannelId: channelId,
		PortId:    portId,
	})
}

type ChannelPort struct {
	Channel string
	Port    string
}

func AllIbcChannelClientStates(chainName string, channelPorts []ChannelPort) ([]ibcchannel.QueryChannelClientStateResponse, error) {

	resp := []ibcchannel.QueryChannelClientStateResponse{}

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcport), grpc.WithInsecure())
	if err != nil {
		return &ibcchannel.QueryChannelClientStateResponse{}, err
	}
	defer func() {
		_ = grpcConn.Close()
	}()

	iq := ibcchannel.NewQueryClient(grpcConn)

	for _, cp := range channelPorts {
		iq := ibcchannel.NewQueryClient(grpcConn)
		res, err := iq.ChannelClientState(context.Background(), &ibcchannel.QueryChannelClientStateRequest{
			ChannelId: cp.Channel,
			PortId:    cp.Port,
		})
		if err != nil {
			return resp, err
		}
		resp = append(resp, res)
	}

	return resp, nil
}

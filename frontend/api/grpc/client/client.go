package client

import (
	"clubmanager/api/grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClubManagerServiceClient(target string) (proto.ClubManagerServiceClient, *grpc.ClientConn, error) {
  conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
  if err != nil {
    return nil, nil, err
  }

  return proto.NewClubManagerServiceClient(conn), conn, nil
}

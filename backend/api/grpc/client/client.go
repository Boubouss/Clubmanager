package client

import (
  "clubmanager/api/grpc/proto"
	"google.golang.org/grpc"
)

func NewClubManagerServiceClient(target string) (proto.ClubManagerServiceClient, *grpc.ClientConn, error) {
  conn, err := grpc.NewClient(target)
  if err != nil {
    return nil, nil, err
  }

  return proto.NewClubManagerServiceClient(conn), conn, nil
}

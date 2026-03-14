package server

import (
	"clubmanager/api/grpc/proto"
	"clubmanager/services/users/models"
	"context"
)


func (s ClubManagerServiceServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {

  user, err := s.usvc.CreateUser(ctx, &models.CreateUserRequest{
    Username: req.Username,
    Email: req.Email,
    Phonenumber: req.Phonenumber,
    Password: req.Password,
  })

  if err != nil {
    return nil, err
  }

  return &proto.CreateUserResponse{
    Id: user.Id.String(),
    Username: user.Username,
    Token: user.Token,
    Errors: user.Errors,
  }, nil
}

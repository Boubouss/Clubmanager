package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)


func (s ClubManagerServiceServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
  user, err := s.usvc.CreateUser(ctx, &dto.CreateUserRequest{
    Username: req.Username,
    Email: req.Email,
    Phonenumber: req.Phonenumber,
    Password: req.Password,
  })

  if err != nil {
    return nil, err
  }

  return &proto.CreateUserResponse{
    User: userProto(user.User),
    Token: user.Token,
    Errors: user.Errors,
  }, nil
}

func (s ClubManagerServiceServer) ReadUser(ctx context.Context, req *proto.ReadUserRequest) (*proto.ReadUserResponse, error) {
  params := make(map[string]any, len(req.Params))
  for k, v := range req.Params {
    params[k] = v
  }

  users, err := s.usvc.ReadUser(ctx, &dto.ReadUserRequest{
    Params: params,
  })

  if err != nil {
    return nil, err
  }

  return &proto.ReadUserResponse{
    Users: arrayUserProto(users.Users),
    Errors: users.Errors,
  }, nil
}


func (s ClubManagerServiceServer) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
  user, err := s.usvc.UpdateUser(ctx, &dto.UpdateUserRequest{
    Email: req.Email,
    Phonenumber: req.Phonenumber,
    Password: req.Password,
  })

  if err != nil {
    return nil, err
  }

  return &proto.UpdateUserResponse{
    User: userProto(user.User),
    Errors: make(map[string]string, 0),
  }, nil
}


func (s ClubManagerServiceServer) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*proto.DeleteUserResponse, error) {
  ok, err := s.usvc.DeleteUser(ctx, req.Token)

  if err != nil {
    return nil, err
  }

  return &proto.DeleteUserResponse{
    Ok: ok,
  }, nil
}

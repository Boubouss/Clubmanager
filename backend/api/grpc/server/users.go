package server

import (
	"clubmanager/api/grpc/proto"
	models "clubmanager/services/users"
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
    User: user.User.Proto(),
    Token: user.Token,
    Errors: user.Errors,
  }, nil
}

func (s ClubManagerServiceServer) ReadUser(ctx context.Context, req *proto.ReadUserRequest) (*proto.ReadUserResponse, error) {
  users, err := s.usvc.ReadUser(ctx, &models.ReadUserRequest{
    Params: req.Params,
  })

  if err != nil {
    return nil, err
  }

  return &proto.ReadUserResponse{
    Users: models.ArrayUserProto(users.Users),
    Errors: users.Errors,
  }, nil
}


func (s ClubManagerServiceServer) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
  user, err := s.usvc.UpdateUser(ctx, &models.UpdateUserRequest{
    Email: req.Email,
    Phonenumber: req.Phonenumber,
    Password: req.Password,
  })

  if err != nil {
    return nil, err
  }

  return &proto.UpdateUserResponse{
    User: user.User.Proto(),
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

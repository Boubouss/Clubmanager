package server

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/adapters/api/grpc/proto"
	"context"
)

func (s ClubManagerServiceServer) CreatePost(ctx context.Context, req *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	res, err := s.psvc.CreatePost(ctx, &dto.CreatePostRequest{
		ClubId:     req.ClubId,
		Title:      req.Title,
		Content:    req.Content,
		Visibility: req.Visibility,
		Status:     req.Status,
	})
	if err != nil {
		return nil, err
	}
	var post *proto.Post
	if res.Post != nil {
		post = postProto(res.Post)
	}
	return &proto.CreatePostResponse{Post: post, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) PublishPost(ctx context.Context, req *proto.PublishPostRequest) (*proto.PublishPostResponse, error) {
	res, err := s.psvc.PublishPost(ctx, &dto.PublishPostRequest{Id: req.Id})
	if err != nil {
		return nil, err
	}
	var post *proto.Post
	if res.Post != nil {
		post = postProto(res.Post)
	}
	return &proto.PublishPostResponse{Post: post, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) UpdatePost(ctx context.Context, req *proto.UpdatePostRequest) (*proto.UpdatePostResponse, error) {
	res, err := s.psvc.UpdatePost(ctx, &dto.UpdatePostRequest{
		Id:         req.Id,
		Title:      req.Title,
		Content:    req.Content,
		Visibility: req.Visibility,
	})
	if err != nil {
		return nil, err
	}
	var post *proto.Post
	if res.Post != nil {
		post = postProto(res.Post)
	}
	return &proto.UpdatePostResponse{Post: post, Errors: res.Errors}, nil
}

func (s ClubManagerServiceServer) DeletePost(ctx context.Context, req *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	ok, err := s.psvc.DeletePost(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.DeletePostResponse{Ok: ok}, nil
}

func (s ClubManagerServiceServer) GetClubPosts(ctx context.Context, req *proto.GetClubPostsRequest) (*proto.GetClubPostsResponse, error) {
	res, err := s.psvc.GetClubPosts(ctx, &dto.GetClubPostsRequest{
		ClubId:   req.ClubId,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	return &proto.GetClubPostsResponse{
		Posts:    arrayPostProto(res.Posts),
		Total:    int32(res.Total),
		Page:     int32(res.Page),
		PageSize: int32(res.PageSize),
		Errors:   res.Errors,
	}, nil
}

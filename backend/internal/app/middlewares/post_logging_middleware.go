package middlewares

import (
	"clubmanager/internal/adapters/api/grpc/dto"
	"clubmanager/internal/app/services"
	"clubmanager/internal/domain/posts"
	"context"
	"fmt"
	"time"
)

type postLoggingService struct {
	next services.PostService
}

func NewPostLoggingService(next services.PostService) services.PostService {
	return &postLoggingService{next: next}
}

func (s *postLoggingService) CreatePost(ctx context.Context, data *dto.CreatePostRequest) (res *dto.CreatePostResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "CreatePost", time.Since(begin), err)
	}(time.Now())
	return s.next.CreatePost(ctx, data)
}

func (s *postLoggingService) PublishPost(ctx context.Context, data *dto.PublishPostRequest) (res *dto.PublishPostResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "PublishPost", time.Since(begin), err)
	}(time.Now())
	return s.next.PublishPost(ctx, data)
}

func (s *postLoggingService) UnpublishPost(ctx context.Context, data *dto.PublishPostRequest) (res *dto.PublishPostResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UnpublishPost", time.Since(begin), err)
	}(time.Now())
	return s.next.UnpublishPost(ctx, data)
}

func (s *postLoggingService) UpdatePost(ctx context.Context, data *dto.UpdatePostRequest) (res *dto.UpdatePostResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "UpdatePost", time.Since(begin), err)
	}(time.Now())
	return s.next.UpdatePost(ctx, data)
}

func (s *postLoggingService) DeletePost(ctx context.Context, id string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "DeletePost", time.Since(begin), err)
	}(time.Now())
	return s.next.DeletePost(ctx, id)
}

func (s *postLoggingService) GetClubPosts(ctx context.Context, data *dto.GetClubPostsRequest) (res *dto.GetClubPostsResponse, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetClubPosts", time.Since(begin), err)
	}(time.Now())
	return s.next.GetClubPosts(ctx, data)
}

func (s *postLoggingService) GetPost(ctx context.Context, id string) (p *posts.Post, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "GetPost", time.Since(begin), err)
	}(time.Now())
	return s.next.GetPost(ctx, id)
}

func (s *postLoggingService) HasPostForEvent(ctx context.Context, eventId string) (ok bool, err error) {
	defer func(begin time.Time) {
		fmt.Printf("=> type: '%s'; took: '%v'; err: '%v'.\n", "HasPostForEvent", time.Since(begin), err)
	}(time.Now())
	return s.next.HasPostForEvent(ctx, eventId)
}

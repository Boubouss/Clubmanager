package dto

import p "clubmanager/internal/domain/posts"

type CreatePostRequest struct {
	ClubId     string
	EventId    string // optional, links the post to an event
	Title      string
	Content    string
	Visibility string
	Status     string // "draft" (default) or "published"
}

func (r CreatePostRequest) Map() map[string]string {
	return map[string]string{
		"event_id":   r.EventId,
		"title":      r.Title,
		"content":    r.Content,
		"visibility": r.Visibility,
		"status":     r.Status,
	}
}

type CreatePostResponse struct {
	Post   *p.Post
	Errors map[string]string
}

type PublishPostRequest struct {
	Id string
}

type PublishPostResponse struct {
	Post   *p.Post
	Errors map[string]string
}

type UpdatePostRequest struct {
	Id         string
	Title      string
	Content    string
	Visibility string
}

func (r UpdatePostRequest) Map() map[string]string {
	m := make(map[string]string)
	if r.Title != "" {
		m["title"] = r.Title
	}
	if r.Content != "" {
		m["content"] = r.Content
	}
	if r.Visibility != "" {
		m["visibility"] = r.Visibility
	}
	return m
}

type UpdatePostResponse struct {
	Post   *p.Post
	Errors map[string]string
}

type GetClubPostsRequest struct {
	ClubId   string
	Page     int
	PageSize int
}

type GetClubPostsResponse struct {
	Posts    []*p.Post
	Total    int
	Page     int
	PageSize int
	Errors   map[string]string
}

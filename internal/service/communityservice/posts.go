package communityservice

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/community"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/pkg/sanitize"
)

func (s *Service) CreatePost(ctx context.Context, userID, body string) (*community.Post, error) {
	body = strings.TrimSpace(sanitize.StripHTML(body))
	if body == "" {
		return nil, apperror.NewValidation("post body is required")
	}

	status, reason := s.classify(ctx, body)

	p := &community.Post{
		ID:        uuid.NewString(),
		UserID:    userID,
		Body:      body,
		Status:    status,
		CreatedAt: s.now(),
	}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, apperror.NewInternal("could not create post").WithCause(err)
	}

	if status == community.StatusUnderReview {
		s.fileSystemReport(ctx, community.TargetTypePost, p.ID, reason)
	}

	return p, nil
}

type ListPostsResult struct {
	Posts []*community.Post
	Total int
}

func (s *Service) ListFeed(ctx context.Context, filter community.ListFilter) (ListPostsResult, error) {
	posts, total, err := s.posts.List(ctx, filter)
	if err != nil {
		return ListPostsResult{}, apperror.NewInternal("could not list posts").WithCause(err)
	}
	return ListPostsResult{Posts: posts, Total: total}, nil
}

func (s *Service) DeletePost(ctx context.Context, userID, postID string) error {
	if err := s.posts.SoftDelete(ctx, postID, userID); err != nil {
		if errors.Is(err, community.ErrPostNotFound) {
			return apperror.NewNotFound("post not found")
		}
		return apperror.NewInternal("could not delete post").WithCause(err)
	}
	return nil
}

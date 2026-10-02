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

func (s *Service) CreateComment(ctx context.Context, userID, postID, body string) (*community.Comment, error) {
	post, err := s.posts.FindByID(ctx, postID)
	if err != nil {
		if errors.Is(err, community.ErrPostNotFound) {
			return nil, apperror.NewNotFound("post not found")
		}
		return nil, apperror.NewInternal("could not add comment").WithCause(err)
	}
	if post.Status == community.StatusRemoved {
		return nil, apperror.NewNotFound("post not found")
	}

	body = strings.TrimSpace(sanitize.StripHTML(body))
	if body == "" {
		return nil, apperror.NewValidation("comment body is required")
	}

	status, reason := s.classify(ctx, body)

	c := &community.Comment{
		ID:        uuid.NewString(),
		PostID:    postID,
		UserID:    userID,
		Body:      body,
		Status:    status,
		CreatedAt: s.now(),
	}
	if err := s.comments.Create(ctx, c); err != nil {
		return nil, apperror.NewInternal("could not add comment").WithCause(err)
	}

	if status == community.StatusUnderReview {
		s.fileSystemReport(ctx, community.TargetTypeComment, c.ID, reason)
	}

	return c, nil
}

type ListCommentsResult struct {
	Comments []*community.Comment
	Total    int
}

func (s *Service) ListComments(ctx context.Context, postID string, filter community.ListFilter) (ListCommentsResult, error) {
	comments, total, err := s.comments.ListByPost(ctx, postID, filter)
	if err != nil {
		return ListCommentsResult{}, apperror.NewInternal("could not list comments").WithCause(err)
	}
	return ListCommentsResult{Comments: comments, Total: total}, nil
}

func (s *Service) DeleteComment(ctx context.Context, userID, commentID string) error {
	if err := s.comments.SoftDelete(ctx, commentID, userID); err != nil {
		if errors.Is(err, community.ErrCommentNotFound) {
			return apperror.NewNotFound("comment not found")
		}
		return apperror.NewInternal("could not delete comment").WithCause(err)
	}
	return nil
}

package dto

import (
	"time"

	"github.com/traumacompendium/backend/internal/domain/community"
)

type CreatePostRequest struct {
	Body string `json:"body" binding:"required,max=5000"`
}

type PostResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewPostResponse(p *community.Post) PostResponse {
	return PostResponse{ID: p.ID, UserID: p.UserID, Body: p.Body, Status: string(p.Status), CreatedAt: p.CreatedAt}
}

func NewPostResponses(posts []*community.Post) []PostResponse {
	out := make([]PostResponse, len(posts))
	for i, p := range posts {
		out[i] = NewPostResponse(p)
	}
	return out
}

type CreateCommentRequest struct {
	Body string `json:"body" binding:"required,max=2000"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewCommentResponse(c *community.Comment) CommentResponse {
	return CommentResponse{ID: c.ID, PostID: c.PostID, UserID: c.UserID, Body: c.Body, Status: string(c.Status), CreatedAt: c.CreatedAt}
}

func NewCommentResponses(comments []*community.Comment) []CommentResponse {
	out := make([]CommentResponse, len(comments))
	for i, c := range comments {
		out[i] = NewCommentResponse(c)
	}
	return out
}

type FileReportRequest struct {
	TargetType string `json:"target_type" binding:"required,oneof=post comment"`
	TargetID   string `json:"target_id" binding:"required"`
	Reason     string `json:"reason" binding:"max=500"`
}

type ResolveReportRequest struct {
	Action string `json:"action" binding:"required,oneof=dismiss remove"`
}

type ReportResponse struct {
	ID               string     `json:"id"`
	TargetType       string     `json:"target_type"`
	TargetID         string     `json:"target_id"`
	ReporterUserID   string     `json:"reporter_user_id,omitempty"`
	Reason           string     `json:"reason,omitempty"`
	Status           string     `json:"status"`
	ResolvedByUserID string     `json:"resolved_by_user_id,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func NewReportResponse(r *community.Report) ReportResponse {
	return ReportResponse{
		ID:               r.ID,
		TargetType:       string(r.TargetType),
		TargetID:         r.TargetID,
		ReporterUserID:   r.ReporterUserID,
		Reason:           r.Reason,
		Status:           string(r.Status),
		ResolvedByUserID: r.ResolvedByUserID,
		ResolvedAt:       r.ResolvedAt,
		CreatedAt:        r.CreatedAt,
	}
}

func NewReportResponses(reports []*community.Report) []ReportResponse {
	out := make([]ReportResponse, len(reports))
	for i, r := range reports {
		out[i] = NewReportResponse(r)
	}
	return out
}

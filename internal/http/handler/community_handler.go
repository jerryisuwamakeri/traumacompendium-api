package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/domain/community"
	apphttp "github.com/traumacompendium/backend/internal/http"
	"github.com/traumacompendium/backend/internal/http/dto"
	"github.com/traumacompendium/backend/internal/http/middleware"
	"github.com/traumacompendium/backend/internal/service/communityservice"
)

type CommunityHandler struct {
	svc *communityservice.Service
}

func NewCommunityHandler(svc *communityservice.Service) *CommunityHandler {
	return &CommunityHandler{svc: svc}
}

// POST /api/v1/community/posts
func (h *CommunityHandler) CreatePost(c *gin.Context) {
	var req dto.CreatePostRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	p, err := h.svc.CreatePost(c.Request.Context(), middleware.UserID(c), req.Body)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewPostResponse(p))
}

// GET /api/v1/community/posts?limit=&offset=
func (h *CommunityHandler) ListFeed(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListFeed(c.Request.Context(), community.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OKWithMeta(c, dto.NewPostResponses(result.Posts), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}

// DELETE /api/v1/community/posts/:id
func (h *CommunityHandler) DeletePost(c *gin.Context) {
	if err := h.svc.DeletePost(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// POST /api/v1/community/posts/:id/comments
func (h *CommunityHandler) CreateComment(c *gin.Context) {
	var req dto.CreateCommentRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	comment, err := h.svc.CreateComment(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.Body)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewCommentResponse(comment))
}

// GET /api/v1/community/posts/:id/comments?limit=&offset=
func (h *CommunityHandler) ListComments(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListComments(c.Request.Context(), c.Param("id"), community.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OKWithMeta(c, dto.NewCommentResponses(result.Comments), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}

// DELETE /api/v1/community/comments/:id
func (h *CommunityHandler) DeleteComment(c *gin.Context) {
	if err := h.svc.DeleteComment(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

// POST /api/v1/community/reports
func (h *CommunityHandler) FileReport(c *gin.Context) {
	var req dto.FileReportRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	r, err := h.svc.FileReport(c.Request.Context(), middleware.UserID(c), community.TargetType(req.TargetType), req.TargetID, req.Reason)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.Created(c, dto.NewReportResponse(r))
}

// GET /api/v1/admin/community/reports?limit=&offset=
func (h *CommunityHandler) ListOpenReports(c *gin.Context) {
	limit, err := apphttp.ParseLimit(c, 20, 100)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	offset, err := apphttp.ParseOffset(c)
	if err != nil {
		apphttp.Fail(c, err)
		return
	}

	result, err := h.svc.ListOpenReports(c.Request.Context(), community.ListFilter{Limit: limit, Offset: offset})
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.OKWithMeta(c, dto.NewReportResponses(result.Reports), gin.H{
		"total": result.Total, "limit": limit, "offset": offset,
	})
}

// POST /api/v1/admin/community/reports/:id/resolve
func (h *CommunityHandler) ResolveReport(c *gin.Context) {
	var req dto.ResolveReportRequest
	if !apphttp.BindJSON(c, &req) {
		return
	}

	err := h.svc.ResolveReport(c.Request.Context(), middleware.UserID(c), c.Param("id"), communityservice.ResolveAction(req.Action))
	if err != nil {
		apphttp.Fail(c, err)
		return
	}
	apphttp.NoContent(c)
}

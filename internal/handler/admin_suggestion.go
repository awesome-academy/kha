package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/service"
)

const (
	adminSuggestionsMenu     = "suggestions"
	adminSuggestionsTitle    = "Đề xuất"
	adminSuggestionsPath     = "/admin/suggestions"
	adminSuggestionsFlashKey = "flash_suggestion"
)

type AdminSuggestionHandler struct {
	suggestionService *service.SuggestionService
	listTmpl          *template.Template
}

func NewAdminSuggestionHandler(suggestionService *service.SuggestionService, funcMap template.FuncMap) *AdminSuggestionHandler {
	layout := "templates/admin/layout.html"
	return &AdminSuggestionHandler{
		suggestionService: suggestionService,
		listTmpl: template.Must(
			template.New("suggestion_list").Funcs(funcMap).ParseFiles(layout, "templates/admin/suggestions/list.html"),
		),
	}
}

func (h *AdminSuggestionHandler) render(c *gin.Context, status int, tmpl *template.Template, data gin.H) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		_ = c.Error(err)
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (h *AdminSuggestionHandler) setFlash(c *gin.Context, t, msg string) {
	c.SetCookie(adminSuggestionsFlashKey, t+"|"+msg, 0, "/", "", false, true)
}

func (h *AdminSuggestionHandler) getFlash(c *gin.Context) *flash {
	val, err := c.Cookie(adminSuggestionsFlashKey)
	if err != nil || val == "" {
		return nil
	}
	c.SetCookie(adminSuggestionsFlashKey, "", -1, "/", "", false, true)
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return &flash{Type: parts[0], Message: parts[1]}
}

func (h *AdminSuggestionHandler) List(c *gin.Context) {
	q := dto.AdminSuggestionListRequest{
		Page:     1,
		PageSize: 15,
		Status:   strings.TrimSpace(c.Query("status")),
		Classify: strings.TrimSpace(c.Query("classify")),
		Search:   strings.TrimSpace(c.Query("search")),
		SortBy:   strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		SortDir:  strings.TrimSpace(c.DefaultQuery("sort_dir", "desc")),
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		q.Page = p
	}

	result, err := h.suggestionService.ListForAdmin(&q)
	if err != nil {
		h.render(c, http.StatusInternalServerError, h.listTmpl, gin.H{
			"Title":      adminSuggestionsTitle,
			"ActiveMenu": adminSuggestionsMenu,
			"Flash":      &flash{Type: flashTypeErr, Message: "Lỗi khi tải danh sách đề xuất: " + err.Error()},
		})
		return
	}

	suggestions, ok := result.Items.([]dto.SuggestionResponse)
	if !ok {
		suggestions = []dto.SuggestionResponse{}
	}

	h.render(c, http.StatusOK, h.listTmpl, gin.H{
		"Title":       adminSuggestionsTitle,
		"ActiveMenu":  adminSuggestionsMenu,
		"Flash":       h.getFlash(c),
		"Suggestions": suggestions,
		"Query":       q,
		"Pagination": paginationData{
			Page:       q.Page,
			TotalPages: result.TotalPages,
			Total:      result.Total,
			Pages:      buildPages(q.Page, result.TotalPages),
		},
	})
}

func (h *AdminSuggestionHandler) UpdateStatus(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID đề xuất không hợp lệ.")
		c.Redirect(http.StatusFound, adminSuggestionsPath)
		return
	}

	status := strings.TrimSpace(c.PostForm("status"))
	note := strings.TrimSpace(c.PostForm("admin_note"))
	req := &dto.AdminUpdateSuggestionStatusRequest{Status: status}
	if note != "" {
		req.AdminNote = &note
	}

	if err := h.suggestionService.UpdateStatusForAdmin(id, req); err != nil {
		h.setFlash(c, flashTypeErr, "Không thể cập nhật đề xuất: "+err.Error())
		c.Redirect(http.StatusFound, adminSuggestionsPath)
		return
	}

	h.setFlash(c, flashTypeOK, fmt.Sprintf("Đã cập nhật đề xuất #%d.", id))
	c.Redirect(http.StatusFound, adminSuggestionsPath)
}

func (h *AdminSuggestionHandler) parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

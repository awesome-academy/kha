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
	flashCookieName = "flash"
	flashTypeOK     = "success"
	flashTypeErr    = "error"
)

type flash struct {
	Type    string
	Message string
}

type categoryFormData struct {
	Name        string
	Slug        string
	Description string
	ImageURL    string
	SortOrder   int
	Status      string
}

type categoryListQuery struct {
	Search string
	Status string
	SortBy string
	Page   int
}

func (q categoryListQuery) URLParams() string {
	parts := []string{}
	if q.Search != "" {
		parts = append(parts, "search="+q.Search)
	}
	if q.Status != "" {
		parts = append(parts, "status="+q.Status)
	}
	if q.SortBy != "" {
		parts = append(parts, "sort_by="+q.SortBy)
	}
	return strings.Join(parts, "&")
}

type paginationData struct {
	Page       int
	TotalPages int
	Total      int64
	Pages      []int
}

// AdminCategoryHandler handles SSR pages for admin category management
type AdminCategoryHandler struct {
	categoryService *service.CategoryService
	listTmpl        *template.Template
	formTmpl        *template.Template
}

// NewAdminCategoryHandler creates a new AdminCategoryHandler and pre-parses templates.
func NewAdminCategoryHandler(categoryService *service.CategoryService, funcMap template.FuncMap) *AdminCategoryHandler {
	layout := "templates/admin/layout.html"
	return &AdminCategoryHandler{
		categoryService: categoryService,
		listTmpl: template.Must(
			template.New("list").Funcs(funcMap).ParseFiles(layout, "templates/admin/categories/list.html"),
		),
		formTmpl: template.Must(
			template.New("form").Funcs(funcMap).ParseFiles(layout, "templates/admin/categories/form.html"),
		),
	}
}

func (h *AdminCategoryHandler) render(c *gin.Context, status int, tmpl *template.Template, data gin.H) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		_ = c.Error(err)
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (h *AdminCategoryHandler) setFlash(c *gin.Context, t, msg string) {
	c.SetCookie(flashCookieName, t+"|"+msg, 0, "/", "", false, true)
}

func (h *AdminCategoryHandler) getFlash(c *gin.Context) *flash {
	val, err := c.Cookie(flashCookieName)
	if err != nil || val == "" {
		return nil
	}
	c.SetCookie(flashCookieName, "", -1, "/", "", false, true)
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return &flash{Type: parts[0], Message: parts[1]}
}

func buildPages(current, total int) []int {
	pages := []int{}
	start := current - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > total {
		end = total
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

// List renders the category list page
func (h *AdminCategoryHandler) List(c *gin.Context) {
	q := categoryListQuery{
		Search: c.Query("search"),
		Status: c.Query("status"),
		SortBy: c.DefaultQuery("sort_by", "sort_order"),
		Page:   1,
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		q.Page = p
	}

	pageSize := 15
	req := &dto.CategoryListRequest{
		Page:     q.Page,
		PageSize: pageSize,
		Status:   q.Status,
		Search:   q.Search,
		SortBy:   q.SortBy,
		SortDir:  "asc",
	}

	result, err := h.categoryService.List(req)
	if err != nil {
		h.render(c, http.StatusInternalServerError, h.listTmpl, gin.H{
			"Title":      "Danh mục",
			"ActiveMenu": "categories",
			"Flash":      &flash{Type: flashTypeErr, Message: "Lỗi khi tải danh sách: " + err.Error()},
		})
		return
	}

	categories, _ := result.Items.([]dto.CategoryResponse)

	pagination := paginationData{
		Page:       q.Page,
		TotalPages: result.TotalPages,
		Total:      result.Total,
		Pages:      buildPages(q.Page, result.TotalPages),
	}

	h.render(c, http.StatusOK, h.listTmpl, gin.H{
		"Title":      "Danh mục",
		"ActiveMenu": "categories",
		"Flash":      h.getFlash(c),
		"Categories": categories,
		"Query":      q,
		"Pagination": pagination,
	})
}

// New renders the create category form
func (h *AdminCategoryHandler) New(c *gin.Context) {
	h.render(c, http.StatusOK, h.formTmpl, gin.H{
		"Title":      "Thêm danh mục",
		"ActiveMenu": "categories",
		"Flash":      h.getFlash(c),
		"Form":       categoryFormData{Status: "active"},
	})
}

// Create handles POST /admin/categories
func (h *AdminCategoryHandler) Create(c *gin.Context) {
	form := h.parseForm(c)

	req := &dto.CreateCategoryRequest{
		Name: form.Name,
	}
	if form.Slug != "" {
		req.Slug = &form.Slug
	}
	if form.Description != "" {
		req.Description = &form.Description
	}
	if form.ImageURL != "" {
		req.ImageURL = &form.ImageURL
	}
	req.SortOrder = &form.SortOrder
	req.Status = &form.Status

	_, err := h.categoryService.Create(req)
	if err != nil {
		errs := h.serviceErrMessages(err)
		h.render(c, http.StatusUnprocessableEntity, h.formTmpl, gin.H{
			"Title":      "Thêm danh mục",
			"ActiveMenu": "categories",
			"Errors":     errs,
			"Form":       form,
		})
		return
	}

	h.setFlash(c, flashTypeOK, fmt.Sprintf("Đã tạo danh mục \"%s\" thành công.", form.Name))
	c.Redirect(http.StatusFound, "/admin/categories")
}

// Edit renders the edit category form
func (h *AdminCategoryHandler) Edit(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		c.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	cat, err := h.categoryService.GetByID(id)
	if err != nil {
		h.setFlash(c, flashTypeErr, "Không tìm thấy danh mục.")
		c.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	h.render(c, http.StatusOK, h.formTmpl, gin.H{
		"Title":      "Sửa danh mục",
		"ActiveMenu": "categories",
		"Flash":      h.getFlash(c),
		"Category":   cat,
	})
}

// Update handles POST /admin/categories/:id/update
func (h *AdminCategoryHandler) Update(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		c.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	form := h.parseForm(c)

	req := &dto.UpdateCategoryRequest{
		Name:      &form.Name,
		SortOrder: &form.SortOrder,
		Status:    &form.Status,
	}
	if form.Slug != "" {
		req.Slug = &form.Slug
	}
	if form.Description != "" {
		req.Description = &form.Description
	}
	if form.ImageURL != "" {
		req.ImageURL = &form.ImageURL
	}

	cat, err := h.categoryService.GetByID(id)
	if err != nil {
		h.setFlash(c, flashTypeErr, "Không tìm thấy danh mục.")
		c.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	_, err = h.categoryService.Update(id, req)
	if err != nil {
		errs := h.serviceErrMessages(err)
		h.render(c, http.StatusUnprocessableEntity, h.formTmpl, gin.H{
			"Title":      "Sửa danh mục",
			"ActiveMenu": "categories",
			"Errors":     errs,
			"Category":   cat,
		})
		return
	}

	h.setFlash(c, flashTypeOK, fmt.Sprintf("Đã cập nhật danh mục \"%s\".", form.Name))
	c.Redirect(http.StatusFound, "/admin/categories")
}

// Delete handles POST /admin/categories/:id/delete
func (h *AdminCategoryHandler) Delete(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		c.Redirect(http.StatusFound, "/admin/categories")
		return
	}

	if err := h.categoryService.Delete(id); err != nil {
		h.setFlash(c, flashTypeErr, "Không thể xoá danh mục: "+err.Error())
	} else {
		h.setFlash(c, flashTypeOK, "Đã xoá danh mục.")
	}

	c.Redirect(http.StatusFound, "/admin/categories")
}

func (h *AdminCategoryHandler) parseIDParam(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func (h *AdminCategoryHandler) parseForm(c *gin.Context) categoryFormData {
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))
	status := c.PostForm("status")
	if status == "" {
		status = "active"
	}
	return categoryFormData{
		Name:        strings.TrimSpace(c.PostForm("name")),
		Slug:        strings.TrimSpace(c.PostForm("slug")),
		Description: strings.TrimSpace(c.PostForm("description")),
		ImageURL:    strings.TrimSpace(c.PostForm("image_url")),
		SortOrder:   sortOrder,
		Status:      status,
	}
}

func (h *AdminCategoryHandler) serviceErrMessages(err error) []string {
	switch {
	case err == service.ErrCategoryNotFound:
		return []string{"Không tìm thấy danh mục."}
	case err == service.ErrSlugAlreadyExists:
		return []string{"Slug này đã được sử dụng, vui lòng chọn slug khác."}
	case err == service.ErrEmptySlug:
		return []string{"Tên danh mục phải chứa ít nhất một ký tự chữ hoặc số."}
	default:
		return []string{"Đã có lỗi xảy ra: " + err.Error()}
	}
}

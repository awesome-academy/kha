package handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/service"
)

const (
	adminUsersMenu      = "users"
	adminUsersTitle     = "Người dùng"
	adminUsersPath      = "/admin/users"
	adminUsersFlashName = "flash_user"
)

type AdminUserHandler struct {
	userService *service.AdminUserService
	listTmpl    *template.Template
	detailTmpl  *template.Template
}

func NewAdminUserHandler(userService *service.AdminUserService, funcMap template.FuncMap) *AdminUserHandler {
	layout := "templates/admin/layout.html"
	return &AdminUserHandler{
		userService: userService,
		listTmpl: template.Must(
			template.New("admin_user_list").Funcs(funcMap).ParseFiles(layout, "templates/admin/users/list.html"),
		),
		detailTmpl: template.Must(
			template.New("admin_user_detail").Funcs(funcMap).ParseFiles(layout, "templates/admin/users/detail.html"),
		),
	}
}

func (h *AdminUserHandler) render(c *gin.Context, status int, tmpl *template.Template, data gin.H) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		_ = c.Error(err)
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (h *AdminUserHandler) setFlash(c *gin.Context, t, msg string) {
	c.SetCookie(adminUsersFlashName, t+"|"+msg, 0, "/", "", false, true)
}

func (h *AdminUserHandler) getFlash(c *gin.Context) *flash {
	val, err := c.Cookie(adminUsersFlashName)
	if err != nil || val == "" {
		return nil
	}
	c.SetCookie(adminUsersFlashName, "", -1, "/", "", false, true)
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return &flash{Type: parts[0], Message: parts[1]}
}

func (h *AdminUserHandler) List(c *gin.Context) {
	q := dto.AdminUserListRequest{
		Page:     1,
		PageSize: 15,
		Search:   strings.TrimSpace(c.Query("search")),
		Status:   strings.TrimSpace(c.Query("status")),
		Role:     strings.TrimSpace(c.Query("role")),
		SortBy:   strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		SortDir:  strings.TrimSpace(c.DefaultQuery("sort_dir", "desc")),
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		q.Page = p
	}

	result, err := h.userService.ListForAdmin(&q)
	if err != nil {
		h.render(c, http.StatusInternalServerError, h.listTmpl, gin.H{
			"Title":      adminUsersTitle,
			"ActiveMenu": adminUsersMenu,
			"Flash":      &flash{Type: flashTypeErr, Message: "Lỗi khi tải danh sách người dùng: " + err.Error()},
		})
		return
	}

	users, ok := result.Items.([]dto.UserResponse)
	if !ok {
		users = []dto.UserResponse{}
	}

	h.render(c, http.StatusOK, h.listTmpl, gin.H{
		"Title":      adminUsersTitle,
		"ActiveMenu": adminUsersMenu,
		"Flash":      h.getFlash(c),
		"Users":      users,
		"Query":      q,
		"Pagination": paginationData{
			Page:       q.Page,
			TotalPages: result.TotalPages,
			Total:      result.Total,
			Pages:      buildPages(q.Page, result.TotalPages),
		},
	})
}

func (h *AdminUserHandler) Detail(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID người dùng không hợp lệ.")
		c.Redirect(http.StatusFound, adminUsersPath)
		return
	}

	user, err := h.userService.GetDetailForAdmin(id)
	if err != nil {
		h.setFlash(c, flashTypeErr, "Không tìm thấy người dùng.")
		c.Redirect(http.StatusFound, adminUsersPath)
		return
	}

	h.render(c, http.StatusOK, h.detailTmpl, gin.H{
		"Title":      fmt.Sprintf("%s #%d", adminUsersTitle, user.ID),
		"ActiveMenu": adminUsersMenu,
		"Flash":      h.getFlash(c),
		"User":       user,
	})
}

func (h *AdminUserHandler) UpdateStatus(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID người dùng không hợp lệ.")
		c.Redirect(http.StatusFound, adminUsersPath)
		return
	}

	status := strings.TrimSpace(c.PostForm("status"))
	if err := h.userService.UpdateStatusForAdmin(id, status, h.currentActorID(c)); err != nil {
		h.setFlash(c, flashTypeErr, h.statusUpdateErrorMessage(err))
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", adminUsersPath, id))
		return
	}

	h.setFlash(c, flashTypeOK, "Đã cập nhật trạng thái người dùng.")
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", adminUsersPath, id))
}

func (h *AdminUserHandler) UpdateRole(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID người dùng không hợp lệ.")
		c.Redirect(http.StatusFound, adminUsersPath)
		return
	}

	role := strings.TrimSpace(c.PostForm("role"))
	if err := h.userService.UpdateRoleForAdmin(id, role); err != nil {
		h.setFlash(c, flashTypeErr, h.roleUpdateErrorMessage(err))
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", adminUsersPath, id))
		return
	}

	h.setFlash(c, flashTypeOK, "Đã cập nhật vai trò người dùng.")
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", adminUsersPath, id))
}

func (h *AdminUserHandler) parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

func (h *AdminUserHandler) currentActorID(c *gin.Context) uint {
	v, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		return 0
	}
	switch id := v.(type) {
	case uint:
		return id
	case int:
		if id > 0 {
			return uint(id)
		}
	}
	return 0
}

func (h *AdminUserHandler) statusUpdateErrorMessage(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, service.ErrAdminUserNotFound):
		return "Không tìm thấy người dùng."
	case errors.Is(err, service.ErrInvalidUserStatus):
		return "Trạng thái người dùng không hợp lệ."
	case errors.Is(err, service.ErrCannotBanAdmin):
		return "Không thể ban tài khoản admin."
	case errors.Is(err, service.ErrCannotBanSelf):
		return "Không thể tự ban chính mình."
	default:
		return "Không thể cập nhật trạng thái: " + err.Error()
	}
}

func (h *AdminUserHandler) roleUpdateErrorMessage(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, service.ErrAdminUserNotFound):
		return "Không tìm thấy người dùng."
	case errors.Is(err, service.ErrInvalidUserRole):
		return "Vai trò người dùng không hợp lệ."
	default:
		return "Không thể cập nhật vai trò: " + err.Error()
	}
}

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
	adminOrdersMenu      = "orders"
	adminOrdersTitle     = "Đơn hàng"
	adminOrderFlash      = "flash_order"
	adminOrdersPath      = "/admin/orders"
	adminOrdersTplList   = "order_list"
	adminOrdersTplDetail = "order_detail"
)

type AdminOrderHandler struct {
	orderService *service.OrderService
	listTmpl     *template.Template
	detailTmpl   *template.Template
}

func NewAdminOrderHandler(orderService *service.OrderService, funcMap template.FuncMap) *AdminOrderHandler {
	layout := "templates/admin/layout.html"
	return &AdminOrderHandler{
		orderService: orderService,
		listTmpl: template.Must(
			template.New(adminOrdersTplList).Funcs(funcMap).ParseFiles(layout, "templates/admin/orders/list.html"),
		),
		detailTmpl: template.Must(
			template.New(adminOrdersTplDetail).Funcs(funcMap).ParseFiles(layout, "templates/admin/orders/detail.html"),
		),
	}
}

func (h *AdminOrderHandler) render(c *gin.Context, status int, tmpl *template.Template, data gin.H) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		_ = c.Error(err)
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}

func (h *AdminOrderHandler) setFlash(c *gin.Context, t, msg string) {
	c.SetCookie(adminOrderFlash, t+"|"+msg, 0, "/", "", false, true)
}

func (h *AdminOrderHandler) getFlash(c *gin.Context) *flash {
	val, err := c.Cookie(adminOrderFlash)
	if err != nil || val == "" {
		return nil
	}
	c.SetCookie(adminOrderFlash, "", -1, "/", "", false, true)
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return &flash{Type: parts[0], Message: parts[1]}
}

func (h *AdminOrderHandler) List(c *gin.Context) {
	q := dto.AdminOrderListRequest{
		Status:   strings.TrimSpace(c.Query("status")),
		FromDate: strings.TrimSpace(c.Query("from_date")),
		ToDate:   strings.TrimSpace(c.Query("to_date")),
		SortBy:   strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		SortDir:  strings.TrimSpace(c.DefaultQuery("sort_dir", "desc")),
		Page:     1,
		PageSize: 15,
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		q.Page = p
	}

	result, err := h.orderService.ListOrdersForAdmin(&q)
	if err != nil {
		h.render(c, http.StatusInternalServerError, h.listTmpl, gin.H{
			"Title":      adminOrdersTitle,
			"ActiveMenu": adminOrdersMenu,
			"Flash":      &flash{Type: flashTypeErr, Message: "Lỗi khi tải danh sách đơn hàng: " + err.Error()},
		})
		return
	}

	orders, ok := result.Items.([]dto.OrderResponse)
	if !ok {
		orders = []dto.OrderResponse{}
	}

	h.render(c, http.StatusOK, h.listTmpl, gin.H{
		"Title":      adminOrdersTitle,
		"ActiveMenu": adminOrdersMenu,
		"Flash":      h.getFlash(c),
		"Orders":     orders,
		"Query":      q,
		"Pagination": paginationData{
			Page:       q.Page,
			TotalPages: result.TotalPages,
			Total:      result.Total,
			Pages:      buildPages(q.Page, result.TotalPages),
		},
	})
}

func (h *AdminOrderHandler) Detail(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID đơn hàng không hợp lệ.")
		c.Redirect(http.StatusFound, adminOrdersPath)
		return
	}

	order, err := h.orderService.GetOrderDetailForAdmin(id)
	if err != nil {
		h.setFlash(c, flashTypeErr, "Không tìm thấy đơn hàng.")
		c.Redirect(http.StatusFound, adminOrdersPath)
		return
	}

	h.render(c, http.StatusOK, h.detailTmpl, gin.H{
		"Title":      fmt.Sprintf("%s #%s", adminOrdersTitle, order.OrderNumber),
		"ActiveMenu": adminOrdersMenu,
		"Flash":      h.getFlash(c),
		"Order":      order,
	})
}

func (h *AdminOrderHandler) UpdateStatus(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID đơn hàng không hợp lệ.")
		c.Redirect(http.StatusFound, adminOrdersPath)
		return
	}

	status := strings.TrimSpace(c.PostForm("status"))
	if err := h.orderService.UpdateOrderStatusForAdmin(id, status); err != nil {
		h.setFlash(c, flashTypeErr, "Không thể cập nhật trạng thái: "+err.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/orders/%d", id))
		return
	}

	h.setFlash(c, flashTypeOK, "Đã cập nhật trạng thái đơn hàng.")
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/orders/%d", id))
}

func (h *AdminOrderHandler) parseIDParam(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

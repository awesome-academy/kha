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

type adminOrderListQuery struct {
	Status   string
	FromDate string
	ToDate   string
	SortBy   string
	SortDir  string
	Page     int
}

func (q adminOrderListQuery) URLParams() string {
	parts := []string{}
	if q.Status != "" {
		parts = append(parts, "status="+q.Status)
	}
	if q.FromDate != "" {
		parts = append(parts, "from_date="+q.FromDate)
	}
	if q.ToDate != "" {
		parts = append(parts, "to_date="+q.ToDate)
	}
	if q.SortBy != "" {
		parts = append(parts, "sort_by="+q.SortBy)
	}
	if q.SortDir != "" {
		parts = append(parts, "sort_dir="+q.SortDir)
	}
	return strings.Join(parts, "&")
}

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
			template.New("order_list").Funcs(funcMap).ParseFiles(layout, "templates/admin/orders/list.html"),
		),
		detailTmpl: template.Must(
			template.New("order_detail").Funcs(funcMap).ParseFiles(layout, "templates/admin/orders/detail.html"),
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
	c.SetCookie("flash_order", t+"|"+msg, 0, "/", "", false, true)
}

func (h *AdminOrderHandler) getFlash(c *gin.Context) *flash {
	val, err := c.Cookie("flash_order")
	if err != nil || val == "" {
		return nil
	}
	c.SetCookie("flash_order", "", -1, "/", "", false, true)
	parts := strings.SplitN(val, "|", 2)
	if len(parts) != 2 {
		return nil
	}
	return &flash{Type: parts[0], Message: parts[1]}
}

func (h *AdminOrderHandler) List(c *gin.Context) {
	q := adminOrderListQuery{
		Status:   strings.TrimSpace(c.Query("status")),
		FromDate: strings.TrimSpace(c.Query("from_date")),
		ToDate:   strings.TrimSpace(c.Query("to_date")),
		SortBy:   strings.TrimSpace(c.DefaultQuery("sort_by", "created_at")),
		SortDir:  strings.TrimSpace(c.DefaultQuery("sort_dir", "desc")),
		Page:     1,
	}
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		q.Page = p
	}

	req := &dto.AdminOrderListRequest{
		Page:     q.Page,
		PageSize: 15,
		Status:   q.Status,
		FromDate: q.FromDate,
		ToDate:   q.ToDate,
		SortBy:   q.SortBy,
		SortDir:  q.SortDir,
	}

	result, err := h.orderService.ListOrdersForAdmin(req)
	if err != nil {
		h.render(c, http.StatusInternalServerError, h.listTmpl, gin.H{
			"Title":      "Đơn hàng",
			"ActiveMenu": "orders",
			"Flash":      &flash{Type: flashTypeErr, Message: "Lỗi khi tải danh sách đơn hàng: " + err.Error()},
		})
		return
	}

	orders, _ := result.Items.([]dto.OrderResponse)

	h.render(c, http.StatusOK, h.listTmpl, gin.H{
		"Title":      "Đơn hàng",
		"ActiveMenu": "orders",
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
		c.Redirect(http.StatusFound, "/admin/orders")
		return
	}

	order, err := h.orderService.GetOrderDetailForAdmin(id)
	if err != nil {
		h.setFlash(c, flashTypeErr, "Không tìm thấy đơn hàng.")
		c.Redirect(http.StatusFound, "/admin/orders")
		return
	}

	h.render(c, http.StatusOK, h.detailTmpl, gin.H{
		"Title":      fmt.Sprintf("Đơn hàng #%s", order.OrderNumber),
		"ActiveMenu": "orders",
		"Flash":      h.getFlash(c),
		"Order":      order,
	})
}

func (h *AdminOrderHandler) UpdateStatus(c *gin.Context) {
	id, ok := h.parseIDParam(c)
	if !ok {
		h.setFlash(c, flashTypeErr, "ID đơn hàng không hợp lệ.")
		c.Redirect(http.StatusFound, "/admin/orders")
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

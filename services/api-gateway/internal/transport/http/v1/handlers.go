package v1

import (
	"fmt"
	"hive/services/api-gateway/internal/domain/shared"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	order    OrderClient
	base     BaseClient
	store    StoreClient
	tracking TrackingClient
	dispatch DispatchClient
}

func NewHandler(
	order OrderClient,
	base BaseClient,
	store StoreClient,
	tracking TrackingClient,
	dispatch DispatchClient,
) *Handler {
	return &Handler{
		order:    order,
		base:     base,
		store:    store,
		tracking: tracking,
		dispatch: dispatch,
	}
}

func (h *Handler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func (h *Handler) CreateOrder(c echo.Context) error {
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid user_id format")
	}

	ctx := c.Request().Context()
	orderInfo, err := h.order.CreateOrder(
		ctx,
		req.UserID,
		req.Items,
		shared.Location{Lat: req.DeliveryLocation.Lat, Lon: req.DeliveryLocation.Lon},
	)
	if err != nil {
		return writeGRPCError(c, err, "failed to create order")
	}

	return c.JSON(http.StatusCreated, CreateOrderResponse{
		OrderID:    orderInfo.ID,
		Status:     string(orderInfo.Status),
		DroneID:    orderInfo.DroneID,
		EtaSeconds: orderInfo.EtaSeconds,
	})
}

func (h *Handler) GetOrder(c echo.Context) error {
	orderID := c.Param("id")
	if _, err := uuid.Parse(orderID); err != nil {
		return jsonError(c, http.StatusBadRequest, "invalid order_id format")
	}

	ctx := c.Request().Context()
	o, err := h.order.GetOrder(ctx, orderID)
	if err != nil {
		return writeGRPCError(c, err, "failed to get order")
	}

	return c.JSON(http.StatusOK, GetOrderResponse{
		OrderID: o.ID,
		UserID:  o.UserID,
		DroneID: o.DroneID,
		Items:   o.Items,
		Status:  string(o.Status),
		Location: Location{
			Lat: o.Location.Lat,
			Lon: o.Location.Lon,
		},
	})
}

func (h *Handler) ListBases(c echo.Context) error {
	offset, limit, err := parsePagination(c)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}
	if limit == 0 {
		return c.JSON(http.StatusOK, ListBasesResponse{Items: []BaseDTO{}})
	}

	ctx := c.Request().Context()
	bases, err := h.base.ListBases(ctx, offset, limit)
	if err != nil {
		return writeGRPCError(c, err, "failed to get bases")
	}

	items := make([]BaseDTO, 0, len(bases))
	for _, b := range bases {
		items = append(items, toBaseDTO(b))
	}
	return c.JSON(http.StatusOK, ListBasesResponse{Items: items})
}

func (h *Handler) ListStores(c echo.Context) error {
	offset, limit, err := parsePagination(c)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}
	if limit == 0 {
		return c.JSON(http.StatusOK, ListStoresResponse{Items: []StoreDTO{}})
	}

	ctx := c.Request().Context()
	stores, err := h.store.ListStores(ctx, offset, limit)
	if err != nil {
		return writeGRPCError(c, err, "failed to get stores")
	}

	items := make([]StoreDTO, 0, len(stores))
	for _, s := range stores {
		items = append(items, toStoreDTO(s))
	}
	return c.JSON(http.StatusOK, ListStoresResponse{Items: items})
}

func (h *Handler) ListDrones(c echo.Context) error {
	offset, limit, err := parsePagination(c)
	if err != nil {
		return jsonError(c, http.StatusBadRequest, err.Error())
	}
	if limit == 0 {
		return c.JSON(http.StatusOK, ListDroneResponse{
			ServerTimeMs: time.Now().UnixMilli(),
			Items:        []DroneDTO{},
		})
	}

	ctx := c.Request().Context()
	drones, err := h.tracking.ListDrones(ctx, offset, limit)
	if err != nil {
		return writeGRPCError(c, err, "failed to get drones")
	}

	items := make([]DroneDTO, 0, len(drones))
	for _, d := range drones {
		a, _ := h.dispatch.GetAssignment(ctx, d.ID)
		items = append(items, toDroneDTO(d, a))
	}
	return c.JSON(http.StatusOK, ListDroneResponse{
		ServerTimeMs: time.Now().UnixMilli(),
		Items:        items,
	})
}

func jsonError(c echo.Context, code int, msg string) error {
	return c.JSON(code, ErrorResponse{Error: msg})
}

func parsePagination(c echo.Context) (int64, int64, error) {
	offsetStr := c.QueryParam("offset")
	limitStr := c.QueryParam("limit")

	var offset int64
	if offsetStr != "" {
		v, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil || v < 0 {
			return 0, 0, fmt.Errorf("invalid offset")
		}
		offset = v
	}

	var limit int64 = 20
	if limitStr != "" {
		v, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit")
		}
		limit = v
	}

	if limit < 0 {
		return 0, 0, fmt.Errorf("invalid limit")
	}
	if limit == 0 {
		return offset, 0, nil
	}
	if limit > 200 {
		limit = 200
	}

	return offset, limit, nil
}

package v1

import (
	"hive/services/api-gateway/internal/domain/shared"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	order    OrderClient
	base     BaseClient
	store    StoreClient
	tracking TrackingClient
}

func NewHandler(
	order OrderClient,
	base BaseClient,
	store StoreClient,
	tracking TrackingClient,
) *Handler {
	return &Handler{
		order:    order,
		base:     base,
		store:    store,
		tracking: tracking,
	}
}

func (h *Handler) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func (h *Handler) CreateOrder(c echo.Context) error {
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user_id format"})
	}

	ctx := c.Request().Context()
	orderInfo, err := h.order.CreateOrder(
		ctx,
		req.UserID,
		req.Items,
		shared.Location{
			Lat: req.DeliveryLocation.Lat,
			Lon: req.DeliveryLocation.Lon,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create order: " + err.Error()})
	}

	resp := CreateOrderResponse{
		OrderID:    orderInfo.ID,
		Status:     string(orderInfo.Status),
		DroneID:    orderInfo.DroneID,
		EtaSeconds: orderInfo.EtaSeconds,
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *Handler) GetOrder(c echo.Context) error {
	orderID := c.Param("id")

	if _, err := uuid.Parse(orderID); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid order_id format"})
	}

	ctx := c.Request().Context()
	o, err := h.order.GetOrder(ctx, orderID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get order: " + err.Error()})
	}

	resp := GetOrderResponse{
		OrderID: o.ID,
		UserID:  o.UserID,
		DroneID: o.DroneID,
		Items:   o.Items,
		Status:  string(o.Status),
		Location: Location{
			Lat: o.Location.Lat,
			Lon: o.Location.Lon,
		},
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListBases(c echo.Context) error {
	offset, _ := strconv.ParseInt(c.QueryParam("offset"), 10, 64)
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		c.JSON(http.StatusOK, ListBasesResponse{
			Bases: []BaseDTO{},
		})
	}

	ctx := c.Request().Context()
	bases, err := h.base.ListBases(ctx, offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get bases: " + err.Error()})
	}
	basesDTO := make([]BaseDTO, 0, len(bases))
	for _, b := range bases {
		basesDTO = append(basesDTO, BaseDTO{
			BaseID:  b.ID,
			Name:    b.Name,
			Address: b.Address,
			Location: Location{
				Lat: b.Location.Lat,
				Lon: b.Location.Lon,
			},
		})
	}
	return c.JSON(http.StatusOK, ListBasesResponse{
		Bases: basesDTO,
	})
}

func (h *Handler) ListStores(c echo.Context) error {
	offset, _ := strconv.ParseInt(c.QueryParam("offset"), 10, 64)
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		c.JSON(http.StatusOK, ListStoresResponse{
			Stores: []StoreDTO{},
		})
	}

	ctx := c.Request().Context()
	stores, err := h.store.ListStores(ctx, offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get stores: " + err.Error()})
	}
	storesDTO := make([]StoreDTO, 0, len(stores))
	for _, s := range stores {
		storesDTO = append(storesDTO, StoreDTO{
			StoreID: s.ID,
			Name:    s.Name,
			Address: s.Address,
			Location: Location{
				Lat: s.Location.Lat,
				Lon: s.Location.Lon,
			},
		})
	}
	return c.JSON(http.StatusOK, ListStoresResponse{
		Stores: storesDTO,
	})
}

func (h *Handler) ListDrones(c echo.Context) error {
	offset, _ := strconv.ParseInt(c.QueryParam("offset"), 10, 64)
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)

	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		c.JSON(http.StatusOK, ListDroneResponse{
			Drones: []DroneDTO{},
		})
	}

	ctx := c.Request().Context()
	drones, err := h.tracking.ListDrones(ctx, offset, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get drones: " + err.Error()})
	}
	dronesDTO := make([]DroneDTO, 0, len(drones))
	for _, d := range drones {
		dronesDTO = append(dronesDTO, DroneDTO{
			DroneID: d.ID,
			Location: Location{
				Lat: d.Location.Lat,
				Lon: d.Location.Lon,
			},
			Battery:             d.Battery,
			SpeedMps:            d.SpeedMps,
			ConsumptionPerMeter: d.ConsumptionPerMeter,
		})
	}
	return c.JSON(http.StatusOK, ListDroneResponse{
		Drones: dronesDTO,
	})
}

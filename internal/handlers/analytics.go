package handlers

import (
	"fmt"
	"time"

	"github.com/NIROOZbx/notification-engine/internal/services"
	"github.com/NIROOZbx/notification-engine/internal/utils"
	"github.com/NIROOZbx/notification-engine/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type AnalyticsHandler struct {
	svc services.AnalyticsService
	log zerolog.Logger
}

func NewAnalyticsHandler(svc services.AnalyticsService, log zerolog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		svc: svc,
		log: log,
	}
}

func (h *AnalyticsHandler) GetAnalytics(c fiber.Ctx) error {
	wid, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "workspace not found")
	}
	workspaceID := utils.UUIDToString(wid)

	start, end, groupBy, err := h.parseAnalyticsParams(c)
	if err != nil {
		return response.BadRequest(c, nil, err.Error())
	}

	res, err := h.svc.GetFullAnalytics(c.Context(), workspaceID, start, end, groupBy)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to fetch analytics")
		return response.InternalServerError(c)
	}

	return response.OK(c, "Analytics fetched successfully", res)
}

func (h *AnalyticsHandler) GetActivityLogs(c fiber.Ctx) error {
	wid, err := utils.GetWID(c)
	if err != nil {
		return response.Unauthorized(c, "workspace not found")
	}
	workspaceID := utils.UUIDToString(wid)

	page := utils.QueryInt32(c, "page", 1)
	pageSize := utils.QueryInt32(c, "page_size", 20)

	var channel, status *string
	if c.Query("channel") != "" {
		cStr := c.Query("channel")
		channel = &cStr
	}
	if c.Query("status") != "" {
		sStr := c.Query("status")
		status = &sStr
	}

	res, err := h.svc.GetActivityLogs(c.Context(), workspaceID, channel, status, page, pageSize)
	if err != nil {
		return response.InternalServerError(c)
	}

	return response.OK(c, "Activity logs fetched successfully", res)
}


func (h *AnalyticsHandler) parseAnalyticsParams(c fiber.Ctx) (time.Time, time.Time, string, error) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	groupBy := c.Query("group_by", "day")
	var start, end time.Time
	var err error

	if startStr == "" {
		start = time.Now().AddDate(0, 0, -30)
	} else {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid start date format (use RFC3339)")
		}
	}

	if endStr == "" {
		end = time.Now()
	} else {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid end date format (use RFC3339)")
		}
	}

	if start.After(end) {
		return time.Time{}, time.Time{}, "", fmt.Errorf("start date cannot be after end date")
	}

	return start, end, groupBy, nil
}

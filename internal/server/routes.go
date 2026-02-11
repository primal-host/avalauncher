package server

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/primal-host/avalauncher/internal/config"
	"github.com/primal-host/avalauncher/internal/manager"
)

func (s *Server) routes() {
	s.echo.GET("/health", s.handleHealth)
	s.echo.GET("/", s.handleDashboard)
	s.echo.GET("/api/status", s.handleStatus)

	// Authenticated API group.
	api := s.echo.Group("/api", s.requireBearer)
	api.POST("/nodes", s.handleCreateNode)
	api.GET("/nodes", s.handleListNodes)
	api.GET("/nodes/:id", s.handleGetNode)
	api.POST("/nodes/:id/start", s.handleStartNode)
	api.POST("/nodes/:id/stop", s.handleStopNode)
	api.DELETE("/nodes/:id", s.handleDeleteNode)
	api.GET("/nodes/:id/logs", s.handleNodeLogs)
	api.GET("/events", s.handleListEvents)
}

// requireBearer is Echo middleware that checks the Authorization header.
func (s *Server) requireBearer(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !s.checkBearer(c) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return next(c)
	}
}

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"version": config.Version,
	})
}

func (s *Server) handleDashboard(c echo.Context) error {
	html := strings.ReplaceAll(dashboardHTML, "{{VERSION}}", config.Version)
	return c.HTML(http.StatusOK, html)
}

func (s *Server) handleStatus(c echo.Context) error {
	authenticated := s.checkBearer(c)
	ctx := c.Request().Context()

	counts := map[string]int64{}
	tables := []string{"hosts", "nodes", "l1s", "events"}
	for _, t := range tables {
		var n int64
		// Table names are hardcoded constants, not user input.
		err := s.db.Pool.QueryRow(ctx, "SELECT count(*) FROM "+t).Scan(&n)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		counts[t] = n
	}

	resp := map[string]any{
		"version": config.Version,
		"counts":  counts,
	}

	if authenticated {
		resp["authenticated"] = true
		nodes, err := s.mgr.ListNodes(ctx)
		if err == nil {
			summaries := make([]manager.NodeSummary, 0, len(nodes))
			for _, n := range nodes {
				summaries = append(summaries, manager.NodeSummary{
					ID:          n.ID,
					Name:        n.Name,
					Image:       n.Image,
					NodeID:      n.NodeID,
					StakingPort: n.StakingPort,
					Status:      n.Status,
				})
			}
			resp["nodes"] = summaries
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateNode(c echo.Context) error {
	var req manager.CreateNodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	node, err := s.mgr.CreateNode(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusCreated, node)
}

func (s *Server) handleListNodes(c echo.Context) error {
	nodes, err := s.mgr.ListNodes(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if nodes == nil {
		nodes = []manager.Node{}
	}
	return c.JSON(http.StatusOK, nodes)
}

func (s *Server) handleGetNode(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	node, err := s.mgr.GetNode(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "node not found"})
	}
	return c.JSON(http.StatusOK, node)
}

func (s *Server) handleStartNode(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	if err := s.mgr.StartNode(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) handleStopNode(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	if err := s.mgr.StopNode(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
}

func (s *Server) handleDeleteNode(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	removeVolumes := c.QueryParam("remove_volumes") == "true"
	if err := s.mgr.DeleteNode(c.Request().Context(), id, removeVolumes); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleNodeLogs(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}
	tail := c.QueryParam("tail")
	reader, err := s.mgr.NodeLogs(c.Request().Context(), id, tail)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	defer reader.Close()

	c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	io.Copy(c.Response().Writer, reader)
	return nil
}

func (s *Server) handleListEvents(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	events, err := s.mgr.ListEvents(c.Request().Context(), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if events == nil {
		events = []manager.Event{}
	}
	return c.JSON(http.StatusOK, events)
}

func (s *Server) checkBearer(c echo.Context) bool {
	if s.adminKey == "" {
		return false
	}
	auth := c.Request().Header.Get("Authorization")
	return strings.TrimPrefix(auth, "Bearer ") == s.adminKey
}

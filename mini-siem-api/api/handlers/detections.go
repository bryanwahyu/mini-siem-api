package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"server-analyst/mini-siem-api/domain"
)

// GetDetections godoc
// @Summary List detections
// @Tags detections
// @Produce json
// @Param severity query string false "Comma separated severities"
// @Param rule_id query string false "Rule ID filter"
// @Param time_from query string false "From timestamp (RFC3339)"
// @Param time_to query string false "To timestamp (RFC3339)"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} handlers.DetectionResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /detections [get]
func GetDetections(repo domain.DetectionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		filter, err := buildDetectionFilter(r)
		if err != nil {
			logger.Warn().Err(err).Msg("invalid detection filter")
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		detections, err := repo.List(r.Context(), filter)
		if err != nil {
			logger.Error().Err(err).Msg("failed to list detections")
			writeError(w, http.StatusInternalServerError, "failed to list detections")
			return
		}

		responses := toDetectionResponses(detections)
		writeJSON(w, http.StatusOK, responses)
	}
}

// GetDetectionByID godoc
// @Summary Get detection detail
// @Tags detections
// @Produce json
// @Param id path int true "Detection ID"
// @Success 200 {object} handlers.DetectionResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /detections/{id} [get]
func GetDetectionByID(repo domain.DetectionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			logger.Warn().Str("id", idParam).Msg("invalid detection id")
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		detection, err := repo.Get(r.Context(), id)
		if err != nil {
			if errorsIsNotFound(err) {
				writeError(w, http.StatusNotFound, "detection not found")
				return
			}
			logger.Error().Err(err).Uint64("id", id).Msg("failed to fetch detection")
			writeError(w, http.StatusInternalServerError, "failed to fetch detection")
			return
		}

		writeJSON(w, http.StatusOK, toDetectionResponses([]*domain.Detection{detection})[0])
	}
}

func buildDetectionFilter(r *http.Request) (domain.DetectionFilter, error) {
	q := r.URL.Query()
	filter := domain.DetectionFilter{}

	if severities := strings.TrimSpace(q.Get("severity")); severities != "" {
		parts := strings.Split(severities, ",")
		for _, part := range parts {
			severity, err := domain.ParseSeverity(part)
			if err != nil {
				return filter, err
			}
			filter.Severity = append(filter.Severity, severity)
		}
	}

	if ruleID := q.Get("rule_id"); ruleID != "" {
		id, err := strconv.ParseUint(ruleID, 10, 64)
		if err != nil {
			return filter, err
		}
		filter.RuleIDs = append(filter.RuleIDs, id)
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return filter, err
		}
		filter.Limit = limit
	} else {
		filter.Limit = 100
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return filter, err
		}
		filter.Offset = offset
	}

	if fromStr := q.Get("time_from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return filter, err
		}
		filter.From = &from
	}

	if toStr := q.Get("time_to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return filter, err
		}
		filter.To = &to
	}

	return filter, nil
}

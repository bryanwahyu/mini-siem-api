package handlers

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"server-analyst/mini-siem-api/domain"
)

var fallbackLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

// GetEvents godoc
// @Summary List events
// @Description Returns persisted events with optional filters
// @Tags events
// @Produce json
// @Param source query string false "Filter by source"
// @Param ip query string false "Filter by source IP"
// @Param time_from query string false "Filter from timestamp (RFC3339)"
// @Param time_to query string false "Filter to timestamp (RFC3339)"
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} handlers.EventResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /events [get]
func GetEvents(repo domain.EventRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)

		filter, err := buildEventFilter(r)
		if err != nil {
			logger.Warn().Err(err).Msg("invalid event filter")
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		events, err := repo.List(r.Context(), filter)
		if err != nil {
			logger.Error().Err(err).Msg("failed to list events")
			writeError(w, http.StatusInternalServerError, "failed to list events")
			return
		}

		logger.Info().Str("action", "list_events").Int("count", len(events)).Msg("events listed")

		responses := make([]EventResponse, 0, len(events))
		for _, evt := range events {
			responses = append(responses, toEventResponse(evt))
		}

		writeJSON(w, http.StatusOK, responses)
	}
}

// GetEventByID godoc
// @Summary Get event detail
// @Description Returns a single event by ID
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} handlers.EventResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /events/{id} [get]
func GetEventByID(repo domain.EventRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)

		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			logger.Warn().Str("id", idParam).Msg("invalid event id")
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		event, err := repo.FindByID(r.Context(), id)
		if err != nil {
			if errorsIsNotFound(err) {
				logger.Info().Str("action", "get_event").Uint64("id", id).Msg("event not found")
				writeError(w, http.StatusNotFound, "event not found")
				return
			}
			logger.Error().Err(err).Uint64("id", id).Msg("failed to fetch event")
			writeError(w, http.StatusInternalServerError, "failed to fetch event")
			return
		}

		logger.Info().Str("action", "get_event").Uint64("id", id).Msg("event fetched")

		writeJSON(w, http.StatusOK, toEventResponse(event))
	}
}

func toEventResponse(event *domain.Event) EventResponse {
	metadata := make(map[string]any)
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			metadata[k] = v
		}
	}
	return EventResponse{
		ID:        event.ID,
		Source:    event.Source,
		IP:        event.IP,
		Message:   event.Message,
		Metadata:  metadata,
		CreatedAt: event.CreatedAt,
	}
}

func buildEventFilter(r *http.Request) (domain.EventFilter, error) {
	q := r.URL.Query()
	filter := domain.EventFilter{
		Source: q.Get("source"),
		IP:     q.Get("ip"),
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return filter, errors.New("invalid limit")
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return filter, errors.New("invalid offset")
		}
		filter.Offset = offset
	}

	if fromStr := q.Get("time_from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			return filter, errors.New("invalid time_from")
		}
		filter.From = &from
	}

	if toStr := q.Get("time_to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			return filter, errors.New("invalid time_to")
		}
		filter.To = &to
	}

	return filter, nil
}

func loggerFromRequest(r *http.Request) zerolog.Logger {
	if ctxLogger := zerolog.Ctx(r.Context()); ctxLogger != nil {
		return ctxLogger.With().Logger()
	}
	return fallbackLogger
}

func errorsIsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

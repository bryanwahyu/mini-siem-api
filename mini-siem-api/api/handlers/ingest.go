package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"server-analyst/mini-siem-api/domain"
	"server-analyst/mini-siem-api/infra/services"
)

// PostEvents godoc
// @Summary Ingest security events
// @Description Accepts single or batched events and runs detection rules
// @Tags events
// @Accept json
// @Produce json
// @Success 202 {object} map[string]any
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /events [post]
func PostEvents(eventService *services.EventService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error().Err(err).Msg("failed to read body")
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		batch, err := MarshalEvents(json.RawMessage(raw))
		if err != nil {
			logger.Warn().Err(err).Msg("invalid event payload")
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		events, err := batch.ToDomain()
		if err != nil {
			logger.Warn().Err(err).Msg("failed to map events")
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx := r.Context()
		var detections []*domain.Detection
		if len(events) == 1 {
			detections, err = eventService.Ingest(ctx, events[0])
		} else {
			detections, err = eventService.IngestBatch(ctx, events)
		}
		if err != nil {
			logger.Error().Err(err).Msg("failed to ingest events")
			writeError(w, http.StatusInternalServerError, "failed to ingest events")
			return
		}

		logger.Info().
			Str("action", "ingest_events").
			Int("events", len(events)).
			Int("detections", len(detections)).
			Msg("events ingested")

		resp := map[string]any{
			"ingested":   len(events),
			"detections": toDetectionResponses(detections),
		}
		writeJSON(w, http.StatusAccepted, resp)
	}
}

func toDetectionResponses(detections []*domain.Detection) []DetectionResponse {
	result := make([]DetectionResponse, 0, len(detections))
	for _, d := range detections {
		metadata := make(map[string]any)
		for k, v := range d.Metadata {
			metadata[k] = v
		}
		result = append(result, DetectionResponse{
			ID:        d.ID,
			EventID:   d.EventID,
			RuleID:    d.RuleID,
			Summary:   d.Summary,
			Severity:  string(d.Severity),
			MatchedAt: d.MatchedAt,
			Metadata:  metadata,
		})
	}
	return result
}

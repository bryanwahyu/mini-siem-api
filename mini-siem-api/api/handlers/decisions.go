package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server-analyst/mini-siem-api/domain"
	"server-analyst/mini-siem-api/infra/services"
)

// GetDecisions godoc
// @Summary List analyst decisions
// @Tags decisions
// @Produce json
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} handlers.DecisionResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /decisions [get]
func GetDecisions(decisionService *services.DecisionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		q := r.URL.Query()
		filter := domain.DecisionFilter{Limit: 100}
		if limitStr := q.Get("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid limit")
				return
			}
			filter.Limit = limit
		}
		if offsetStr := q.Get("offset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid offset")
				return
			}
			filter.Offset = offset
		}

		decisions, err := decisionService.List(r.Context(), filter)
		if err != nil {
			logger.Error().Err(err).Msg("failed to list decisions")
			writeError(w, http.StatusInternalServerError, "failed to list decisions")
			return
		}

		responses := make([]DecisionResponse, 0, len(decisions))
		for _, decision := range decisions {
			responses = append(responses, toDecisionResponse(decision))
		}

		writeJSON(w, http.StatusOK, responses)
	}
}

// PostDecisions godoc
// @Summary Record a decision or action
// @Tags decisions
// @Accept json
// @Produce json
// @Success 201 {object} handlers.DecisionResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /decisions [post]
func PostDecisions(decisionService *services.DecisionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		var payload DecisionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid payload")
			return
		}

		decision, err := payload.ToDomain()
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := decisionService.Create(r.Context(), decision); err != nil {
			logger.Error().Err(err).Msg("failed to create decision")
			writeError(w, http.StatusInternalServerError, "failed to create decision")
			return
		}

		writeJSON(w, http.StatusCreated, toDecisionResponse(decision))
	}
}

func toDecisionResponse(decision *domain.Decision) DecisionResponse {
	return DecisionResponse{
		ID:          decision.ID,
		DetectionID: decision.DetectionID,
		Action:      string(decision.Action),
		Target:      decision.Target,
		Reason:      decision.Reason,
		Notes:       decision.Notes,
		CreatedAt:   decision.CreatedAt,
	}
}

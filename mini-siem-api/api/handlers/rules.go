package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"server-analyst/mini-siem-api/domain"
	"server-analyst/mini-siem-api/infra/services"
)

// GetRules godoc
// @Summary List detection rules
// @Tags rules
// @Produce json
// @Success 200 {array} handlers.RuleResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /rules [get]
func GetRules(ruleService *services.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		rules, err := ruleService.List(r.Context())
		if err != nil {
			logger.Error().Err(err).Msg("failed to list rules")
			writeError(w, http.StatusInternalServerError, "failed to list rules")
			return
		}

		responses := make([]RuleResponse, 0, len(rules))
		for _, rule := range rules {
			responses = append(responses, toRuleResponse(rule))
		}

		writeJSON(w, http.StatusOK, responses)
	}
}

// PostRules godoc
// @Summary Create a detection rule
// @Tags rules
// @Accept json
// @Produce json
// @Success 201 {object} handlers.RuleResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /rules [post]
func PostRules(ruleService *services.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		var payload RuleRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Warn().Err(err).Msg("invalid rule payload")
			writeError(w, http.StatusBadRequest, "invalid payload")
			return
		}

		rule, err := payload.ToDomain()
		if err != nil {
			logger.Warn().Err(err).Msg("invalid rule request")
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := ruleService.Create(r.Context(), rule); err != nil {
			logger.Error().Err(err).Msg("failed to create rule")
			writeError(w, http.StatusInternalServerError, "failed to create rule")
			return
		}

		writeJSON(w, http.StatusCreated, toRuleResponse(rule))
	}
}

// PatchRuleActivation godoc
// @Summary Toggle rule activation
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 204 ""
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /rules/{id} [patch]
func PatchRuleActivation(ruleService *services.RuleService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFromRequest(r)
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			logger.Warn().Str("id", idParam).Msg("invalid rule id")
			writeError(w, http.StatusBadRequest, "invalid id")
			return
		}

		var payload RuleActivationRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Warn().Err(err).Msg("invalid activation payload")
			writeError(w, http.StatusBadRequest, "invalid payload")
			return
		}

		if err := ruleService.SetActivation(r.Context(), uint64(id), payload.Active); err != nil {
			logger.Error().Err(err).Uint64("rule_id", uint64(id)).Msg("failed to update rule activation")
			writeError(w, http.StatusInternalServerError, "failed to update rule")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func toRuleResponse(rule *domain.Rule) RuleResponse {
	tags := make([]string, len(rule.Tags))
	copy(tags, rule.Tags)
	return RuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Pattern:     rule.Pattern,
		Description: rule.Description,
		Severity:    string(rule.Severity),
		Tags:        tags,
		Active:      rule.Active,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

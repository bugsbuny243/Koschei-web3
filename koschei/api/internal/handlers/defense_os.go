package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"koschei/api/internal/defense"
	"koschei/api/internal/router"
)

func (h *Handler) OwnerDefenseArtifacts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, err := defense.ListArtifacts(r.Context(), h.DB, r.URL.Query().Get("program_id"), r.URL.Query().Get("network"), limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "artifact_list_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "artifacts": items})
	case http.MethodPost:
		var input defense.ArtifactInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_artifact"})
			return
		}
		input.CreatedBy = "owner"
		item, err := defense.StoreArtifact(r.Context(), h.DB, input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "artifact_rejected", "details": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "artifact": item})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) OwnerDefenseKnowledge(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		framework := strings.TrimSpace(r.URL.Query().Get("framework"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		embedding := []float64{}
		if query != "" && envBool("KOSCHEI_DEFENSE_KNOWLEDGE_EMBEDDINGS_ENABLED", false) {
			if result, err := router.Embed(r.Context(), router.EmbedRequest{Input: query}); err == nil {
				embedding = result.Embedding
			}
		}
		items, err := defense.SearchKnowledge(r.Context(), h.DB, query, framework, embedding, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "knowledge_search_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "documents": items, "embedding_used": len(embedding) > 0})
	case http.MethodPost:
		var input defense.KnowledgeInput
		if err := decodeJSON(r, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_knowledge_document"})
			return
		}
		input.CreatedBy = "owner"
		if envBool("KOSCHEI_DEFENSE_KNOWLEDGE_EMBEDDINGS_ENABLED", false) {
			if result, err := router.Embed(r.Context(), router.EmbedRequest{Input: input.Title + "\n" + input.Body}); err == nil {
				input.Embedding, input.EmbeddingModel = result.Embedding, result.Model
			}
		}
		doc, err := defense.StoreKnowledge(r.Context(), h.DB, input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "knowledge_rejected", "details": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "document": doc})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) OwnerDefenseLab(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method_not_allowed"})
		return
	}
	var input struct {
		Action      string `json:"action"`
		ArtifactRef string `json:"artifact_ref"`
	}
	if err := decodeJSON(r, &input); err != nil || strings.TrimSpace(input.Action) != "analyze" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_static_analysis_request"})
		return
	}
	artifact, err := defense.LoadArtifact(r.Context(), h.DB, input.ArtifactRef)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "artifact_not_found"})
		return
	}
	report, err := defense.AnalyzeArtifact(artifact)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "analysis_failed"})
		return
	}
	if err := defense.PersistLabReport(r.Context(), h.DB, report); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "analysis_persist_failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "report": report})
}

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/reddec/view"

	"github.com/pikocloud/pikomirror/internal/dbo"
)

const pageSize = 20

func Web(store *dbo.Queries) http.Handler {
	router := chi.NewMux()

	router.Get("/", Handler(Index, func(ctx context.Context, req *http.Request) (IndexArgs, error) {
		topEndpoints, err := store.EndpointListTop(ctx, 10)
		if err != nil {
			return IndexArgs{}, fmt.Errorf("list top endpoints: %w", err)
		}

		return IndexArgs{
			TopEndpoints: topEndpoints,
		}, nil
	}))

	router.Get("/requests/", Handler(RequestTable, func(ctx context.Context, req *http.Request) (HeadlineArgs, error) {
		var page = 1
		if v, err := strconv.Atoi(req.FormValue("page")); err == nil {
			page = v
		}

		list, err := store.HeadlineList(ctx, dbo.HeadlineListParams{
			Offset: int32((page - 1) * pageSize),
			Limit:  pageSize,
		})
		if err != nil {
			return HeadlineArgs{}, fmt.Errorf("list requests: %w", err)
		}

		return HeadlineArgs{
			Records:  list,
			Page:     page,
			PageSize: pageSize,
			Term:     "All requests",
		}, nil
	}))

	router.Get("/requests/endpoint/{domain}/{method}/{path}/", Handler(RequestTable, func(ctx context.Context, req *http.Request) (HeadlineArgs, error) {
		var page = 1
		if v, err := strconv.Atoi(req.FormValue("page")); err == nil {
			page = v
		}
		domain, _ := url.PathUnescape(chi.URLParam(req, "domain"))
		method, _ := url.PathUnescape(chi.URLParam(req, "method"))
		path, _ := url.PathUnescape(chi.URLParam(req, "path"))

		list, err := store.HeadlineListByEndpoint(ctx, dbo.HeadlineListByEndpointParams{
			Domain: domain,
			Method: strings.ToUpper(method),
			Path:   path,
			Offset: int32((page - 1) * pageSize),
			Limit:  pageSize,
		})
		if err != nil {
			return HeadlineArgs{}, fmt.Errorf("list requests: %w", err)
		}

		return HeadlineArgs{
			Records:  list,
			Page:     page,
			PageSize: pageSize,
			Term:     "Endpoint: " + domain + " " + method + " " + path,
		}, nil
	}))

	router.Get("/requests/id/{id}/", Handler(RequestView, func(ctx context.Context, req *http.Request) (dbo.Request, error) {
		id, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			return dbo.Request{}, fmt.Errorf("parse ID: %w", err)
		}
		return store.RequestGet(ctx, id)
	}))

	router.Get("/requests/id/{id}/content", func(writer http.ResponseWriter, request *http.Request) {
		// get content as downloadable asset
		id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
		if err != nil {
			slog.Error("failed parse ID", "error", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		req, err := store.RequestGet(request.Context(), id)
		if err != nil {
			slog.Error("failed get request", "error", err, "id", id)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var headers http.Header
		if err := json.Unmarshal(req.Headers, &headers); err != nil {
			slog.Error("failed parse headers", "error", err, "id", id)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var filename = strconv.FormatInt(req.ID, 10)
		if req.RequestID != "" {
			filename += "-" + req.RequestID
		}

		if ct := headers.Get("Content-Type"); ct != "" {
			writer.Header().Add("Content-Type", ct)
		} else {
			writer.Header().Add("Content-Type", "application/octet-stream")
		}

		if rid := req.RequestID; rid != "" {
			writer.Header().Set("X-Request-Id", rid)
		}

		writer.Header().Set("Content-Length", strconv.Itoa(len(req.Content)))
		writer.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		writer.Header().Set("Last-Modified", req.CreatedAt.UTC().Format(time.RFC1123))

		writer.WriteHeader(http.StatusOK)

		if _, err := writer.Write(req.Content); err != nil {
			slog.Error("failed writer content", "error", err, "id", id)
		}
	})

	router.Get("/endpoints/", Handler(EndpointTable, func(ctx context.Context, req *http.Request) (EndpointArgs, error) {
		var page = 1
		if v, err := strconv.Atoi(req.FormValue("page")); err == nil {
			page = v
		}

		list, err := store.EndpointList(ctx, dbo.EndpointListParams{
			Offset: int32((page - 1) * pageSize),
			Limit:  pageSize,
		})
		if err != nil {
			return EndpointArgs{}, fmt.Errorf("list endpoints: %w", err)
		}

		return EndpointArgs{
			Records:  list,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}))

	return router

}

func Handler[T any](v *view.View[T], handler func(ctx context.Context, req *http.Request) (T, error)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		data, err := handler(ctx, request)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed execute", "error", err)
			return
		}
		if err := v.Execute(writer, data); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed render result", "error", err)
			return
		}
	}
}

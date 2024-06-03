package web

import (
	"github.com/pikocloud/pikomirror/internal/dbo"
)

var (
	Index        = View[IndexArgs]("views/index.gohtml")
	RequestTable = View[HeadlineArgs]("views/requests/list.gohtml")
	RequestView  = View[dbo.Request]("views/requests/get.gohtml")

	EndpointTable = View[EndpointArgs]("views/endpoints/list.gohtml")
)

type IndexArgs struct {
	TopEndpoints []dbo.Endpoint
}

type HeadlineArgs struct {
	Records  []dbo.Headline
	Page     int
	PageSize int
	Term     string
}

type EndpointArgs struct {
	Records  []dbo.Endpoint
	Page     int
	PageSize int
}

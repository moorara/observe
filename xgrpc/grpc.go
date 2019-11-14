package xgrpc

import (
	"context"
	"regexp"

	"google.golang.org/grpc"
)

const (
	requestIDKey  = "request-id"
	clientNameKey = "client-name"
)

var methodRegex = regexp.MustCompile(`(/|\.)`)

func parseMethod(fullMethod string) (string, string, string, bool) {
	// fullMethod should have the form /package.service/method
	subs := methodRegex.Split(fullMethod, 4)
	if len(subs) != 4 {
		return "", "", "", false
	}

	return subs[1], subs[2], subs[3], true
}

// filter can be used for filtering a package, a service, or a method from being observed.
type filter struct {
	pkg     string
	service string
	method  string
}

func (f *filter) matches(pkg, service, method string) bool {
	return (f.pkg == pkg && f.service == "" && f.method == "") ||
		(f.pkg == pkg && f.service == service && f.method == "") ||
		(f.pkg == pkg && f.service == service && f.method == method)
}

type xServerStream struct {
	grpc.ServerStream
	context context.Context
}

func (s *xServerStream) Context() context.Context {
	if s.context == nil {
		return s.ServerStream.Context()
	}

	return s.context
}

// ServerStreamWithContext return new grpc.ServerStream with a new context.
func ServerStreamWithContext(stream grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	if ss, ok := stream.(*xServerStream); ok {
		return ss
	}

	return &xServerStream{
		ServerStream: stream,
		context:      ctx,
	}
}

# xhttp

This package provides utilities for HTTP servers and clients.

| Item                     | Description                                                                                       |
|--------------------------|---------------------------------------------------------------------------------------------------|
| `xhttp.Error`            | An `error` type capturing context and information about a failed http request.                    |
| `xhttp.ResponseWriter`   | An implementation of standard `xhttp.ResponseWriter` for recording status code.                   |
| `xhttp.ClientMiddleware` | A client-side middleware providing wrappers for logging, metrics, tracing, etc.                   |
| `xhttp.ServerMiddleware` | A server-side middleware providing wrappers for http handlers for logging, metrics, tracing, etc. |

## Quick Start

You can see an example of using the middleware [here](./example).

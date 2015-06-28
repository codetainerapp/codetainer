package main

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

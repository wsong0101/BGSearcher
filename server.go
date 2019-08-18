package main

import (
	"html/template"
	"io"
	"net/http"
	"runtime"

	"github.com/labstack/echo"

	"bgsearcher.com/api"
	"bgsearcher.com/cloud"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	runtime.GOMAXPROCS(1)

	cloud.InitializeCloud()
	api.UpdateNewArrivals(1800) // every 30 min

	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = renderer

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", map[string]interface{}{
			"content": "main",
		})
	}).Name = "main"

	e.POST("/search", func(c echo.Context) error {
		query := c.QueryParam("query")
		result := api.Search(query)
		return c.JSON(http.StatusOK, result)
	})

	e.POST("/hits", func(c echo.Context) error {
		result := cloud.GetHits()
		return c.JSON(http.StatusOK, result)
	})

	e.POST("/newarrivals", func(c echo.Context) error {
		result := api.GetNewArrivalsFromCache()
		return c.JSON(http.StatusOK, result)
	})

	e.File("/favicon.ico", "favicon.ico")

	// admin
	e.GET("/admin", func(c echo.Context) error {
		return c.Render(http.StatusOK, "admin.html", map[string]interface{}{
			"name": "Dolly!",
		})
	}).Name = "foobar"

	e.POST("/remove", func(c echo.Context) error {
		passwd := c.QueryParam("passwd")
		word := c.QueryParam("word")
		result := cloud.RemoveHistory(word, passwd)
		return c.JSON(http.StatusOK, result)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

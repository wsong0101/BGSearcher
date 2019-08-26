package main

import (
	"html/template"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/labstack/echo"

	"bgsearcher.com/api"
	"bgsearcher.com/cloud"
	"bgsearcher.com/crawl"
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

//PageData is a struct for rendering info
type PageData struct {
	Content   string
	ShopInfos []crawl.ShopInfo
}

func main() {
	runtime.GOMAXPROCS(1)

	cloud.InitializeCloud()
	api.LoadNewArrivalsFromCloud()

	var refreshDuration = 30 * time.Minute
	go api.UpdateNewArrivals(refreshDuration) // every 30 min

	shopInfos := api.GetShopInfos()

	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = renderer

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", &PageData{
			Content:   "main",
			ShopInfos: shopInfos,
		})
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", &PageData{
			Content:   "search",
			ShopInfos: shopInfos,
		})
	})

	e.GET("/new-arrivals", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", PageData{"new-arrivals", nil})
	})

	// admin
	e.GET("/admin", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", PageData{"admin", nil})
	})

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
	e.File("/ror.xml", "ror.xml")

	e.POST("/remove", func(c echo.Context) error {
		passwd := c.QueryParam("passwd")
		word := c.QueryParam("word")
		result := cloud.RemoveHistory(word, passwd)
		return c.JSON(http.StatusOK, result)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/yaml.v2"

	"bgsearcher.com/api"
	"bgsearcher.com/cloud"
	"bgsearcher.com/crawl"
	"bgsearcher.com/ranking"
)

type Config struct {
	Server struct {
		Port   string `yaml:"port"`
		Passwd string `yaml:"passwd"`
	} `yaml:"server"`
}

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

type pageData struct {
	Content   string
	ShopInfos []crawl.ShopInfo
	Weekly    []ranking.QueryCount
	Monthly   []ranking.QueryCount
}

func main() {
	runtime.GOMAXPROCS(1)

	cloud.InitializeCloud()
	api.LoadNewArrivalsFromCloud()
	ranking.InitRanking()

	var refreshDuration = 30 * time.Minute
	go api.UpdateNewArrivals(refreshDuration) // every 30 min

	shopInfos := api.GetShopInfos()

	e := echo.New()
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = renderer

	f, err := os.Open("config.yaml")
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer f.Close()

	var config Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", &pageData{
			Content:   "main",
			ShopInfos: shopInfos,
			Monthly:   ranking.GetMonthlyRank(),
			Weekly:    ranking.GetWeeklyRank(),
		})
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", &pageData{
			Content:   "search",
			ShopInfos: shopInfos,
		})
	})

	e.GET("/new-arrivals", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", pageData{"new-arrivals", nil, nil, nil})
	})

	// admin
	e.GET("/admin", func(c echo.Context) error {
		return c.Render(http.StatusOK, "base.html", pageData{"admin", nil, nil, nil})
	})

	e.POST("/search", func(c echo.Context) error {
		query := c.QueryParam("query")
		result := api.Search(query)
		return c.JSON(http.StatusOK, result)
	})

	e.POST("/ranks", func(c echo.Context) error {
		result := ranking.GetMonthlyRank()
		result = append(result, ranking.GetWeeklyRank()...)
		return c.JSON(http.StatusOK, result)
	})

	e.POST("/hourly", func(c echo.Context) error {
		result := ranking.GetHourlyRank()
		return c.JSON(http.StatusOK, result)
	})

	e.POST("/newarrivals", func(c echo.Context) error {
		result := api.GetNewArrivalsFromCache()
		return c.JSON(http.StatusOK, result)
	})

	e.File("/favicon.ico", "favicon.ico")
	e.File("/ror.xml", "ror.xml")
	e.File("/ads.txt", "ads.txt")

	e.POST("/remove", func(c echo.Context) error {
		passwd := c.QueryParam("passwd")
		if passwd != config.Server.Passwd {
			return c.JSON(http.StatusOK, "wrong password")
		}
		word := c.QueryParam("word")
		ranking.RemoveQuery(word)
		return c.JSON(http.StatusOK, "ok")
	})

	e.Logger.Fatal(e.Start(":" + config.Server.Port))
}

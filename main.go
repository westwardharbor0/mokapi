package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	definitionsPath string
	serverHost      string
	serverPort      int
	checkInterval   time.Duration
	debug           bool

	log *zap.Logger

	apiDefinitions  *Definitions
	fileDefinitions *Definitions
)

func parseArgs() {
	flag.StringVar(
		&definitionsPath,
		"definitions-path",
		"./fileDefinitions",
		"Path to folder containing the fileDefinitions of endpoints",
	)
	flag.StringVar(
		&serverHost,
		"host",
		"localhost",
		"Host we start service on",
	)
	flag.DurationVar(
		&checkInterval,
		"check-interval",
		5*time.Second,
		"Interval to check fileDefinitions for changes",
	)
	flag.IntVar(
		&serverPort,
		"port",
		8080,
		"Post we start service on",
	)
	flag.BoolVar(
		&debug,
		"debug",
		false,
		"Debug logging toggle",
	)
	flag.Parse()
}

func main() {
	parseArgs()

	var logError error
	gin.SetMode(gin.ReleaseMode)
	if debug {
		log, logError = zap.NewDevelopment()
	} else {
		log, logError = zap.NewProduction()
		gin.SetMode(gin.ReleaseMode)
	}
	if logError != nil {
		panic(logError.Error())
	}
	log.Info(
		"Starting the MokAPI service",
		zap.String("host", serverHost),
		zap.Int("port", serverPort),
	)

	fileDefinitions = &Definitions{Path: definitionsPath}
	apiDefinitions = &Definitions{}
	apiDefinitions.Endpoints = make(map[string]*Definition)

	// Only load and watch if we have path to use.
	if strings.TrimSpace(definitionsPath) != "" {
		log.Info("Scanning the path for fileDefinitions", zap.String("path", definitionsPath))
		if err := fileDefinitions.Load(); err != nil {
			log.Fatal("Failed to fileDefinitions", zap.Error(err))
		}
		go watchDefinitions()
	}

	engine := gin.Default()
	engine.Use(reloadMiddleware)
	// Prepare the core endpoints.
	mockAPIEndpoints := engine.Group("/mokapi")
	mockAPIEndpoints.GET("/stats", func(c *gin.Context) {
		c.JSON(200, map[string]any{
			"api_definitions":  len(apiDefinitions.Endpoints),
			"file_definitions": len(fileDefinitions.Endpoints),
		})
	})
	// Handle adding new mock endpoints from request.
	mockAPIEndpoints.POST("/add", func(c *gin.Context) {
		var definition Definition
		err := c.ShouldBindJSON(&definition)
		if err != nil {
			c.JSON(500, err.Error())
		}
		definition.Method = strings.ToUpper(definition.Method)
		// We don't care about duplicates, we overwrite it.
		_ = apiDefinitions.Add(&definition)
	})

	panic(engine.Run(fmt.Sprintf("%s:%d", serverHost, serverPort)))
}

// reloadMiddleware handles requests if found in definitions.
// We use middleware to dynamically add and remove endpoints.
func reloadMiddleware(c *gin.Context) {
	if def, err := getDefinition(c); err == nil {
		c.JSON(def.ResponseStatusCode, def.ResponsePayload)
		c.Done()
	}
}

// getDefinition returns endpoint definition based on the request URI.
// We first search for paths with query params and then only path.
func getDefinition(c *gin.Context) (*Definition, error) {
	keyParams := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.RequestURI)
	keyHollow := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	if def, exists := apiDefinitions.Endpoints[keyParams]; exists {
		return def, nil
	}
	if def, exists := fileDefinitions.Endpoints[keyParams]; exists {
		return def, nil
	}

	if def, exists := apiDefinitions.Endpoints[keyHollow]; exists {
		return def, nil
	}
	if def, exists := fileDefinitions.Endpoints[keyHollow]; exists {
		return def, nil
	}

	return nil, errors.New("failed to find definition")
}

// watchDefinitions detects file changes and triggers endpoint refresh if needed.
func watchDefinitions() {
	for {
		for _, def := range fileDefinitions.Endpoints {
			changed, err := def.Changed()
			if err != nil {
				log.Error("Failed to check definition stats", zap.Error(err))
				continue
			}
			if changed {
				log.Info(fmt.Sprintf("Definition in %s has changed, reloading", def.Path))
				if err := fileDefinitions.Load(); err != nil {
					log.Error("Failed to reload fileDefinitions", zap.Error(err))
				}
			}
		}
		time.Sleep(checkInterval)
	}
}

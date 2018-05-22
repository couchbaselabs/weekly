package main

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func getBuilds(c *gin.Context) {
	builds, err := ds.getBuilds()
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.IndentedJSON(200, builds)
}

func updateBuildStatus(c *gin.Context) {
	var status Status
	if err := c.BindJSON(&status); err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}
	err := ds.updateStatus(status)
	if err != nil {
		c.AbortWithError(500, err)
	}
}

func getBuildStatus(c *gin.Context) {
	build := c.Param("build")
	if build == "" {
		c.AbortWithError(400, errors.New("missing arguments"))
		return
	}
	status, err := ds.getBuildStatus(build)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.IndentedJSON(200, status)
}

func httpEngine() *gin.Engine {
	router := gin.Default()

	router.StaticFile("/", "./app/index.html")
	router.Static("/static", "./app")

	rg := router.Group("/api/v1")

	rg.GET("builds", getBuilds)

	rg.POST("status", updateBuildStatus)

	rg.GET("status/:build", getBuildStatus)

	return router
}

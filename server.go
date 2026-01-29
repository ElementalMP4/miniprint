package main

import (
	"fmt"
	"io"
	"miniprint/printer"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
		}
	}
}

func serve() error {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(ErrorHandler())

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/print", func(ctx *gin.Context) {
		jsonData, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.Error(err)
			return
		}

		format, err := printer.ReceiptFormatFromJson(string(jsonData))
		if err != nil {
			ctx.Error(err)
			return
		}

		if err := printerInterface.QueueReceiptFormat(format); err != nil {
			ctx.Error(fmt.Errorf("Error printing receipt: %v\n", err))
			return
		}

		if format.Settings.NoCut {
			printerInterface.ExecuteWithoutCut()
		} else {
			printerInterface.Execute()
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "Printed successfully",
		})
	})

	r.POST("/cut", func(ctx *gin.Context) {
		printerInterface.Cut()
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Paper has been cut",
		})
	})

	r.POST("/reset", func(ctx *gin.Context) {
		printerInterface.HardResetPrinter()
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Printer has been reset",
		})
	})

	if err := r.Run(); err != nil {
		return fmt.Errorf("failed to run server: %v", err)
	}

	return nil
}

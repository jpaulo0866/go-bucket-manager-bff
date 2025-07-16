package api

import (
	"context"
	"encoding/json"
	"go-bucket-manager-bff/internal/auth"
	"go-bucket-manager-bff/internal/config"
	"go-bucket-manager-bff/internal/storage"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "bff_session"

type googleUserInfo struct {
	Email string `json:"email"`
}

func RegisterAuthRoutes(r *gin.Engine, oauth *auth.OAuthManager, jwt *auth.JWTManager, cfg *config.Config) {
	r.GET("/oauth2/authorization/google", func(c *gin.Context) {
		url := oauth.GetAuthURL("state")
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/login/oauth2/code/google", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
			return
		}
		token, err := oauth.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "oauth2 exchange failed"})
			return
		}
		// Get user info from Google
		client := oauth.Config().Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil || resp.StatusCode != 200 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user info"})
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var userInfo googleUserInfo
		if err := json.Unmarshal(body, &userInfo); err != nil || userInfo.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user info"})
			return
		}
		// Set session cookie (secure, httpOnly)
		// This is an insecure Cookie, for production, enable secure and protect with a domain
		c.SetCookie(sessionCookieName, userInfo.Email, 3600, "/", "", false, true)
		c.Redirect(http.StatusFound, "/session")
	})

	r.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Bucket Manager Bff"})
	})

	r.GET("/session", func(c *gin.Context) {
		session, err := c.Cookie(sessionCookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"bff_session": session})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/auth/token", func(c *gin.Context) {
		email, err := c.Cookie(sessionCookieName)
		if err != nil || email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no session"})
			return
		}
		token, err := jwt.Generate(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token":     token,
			"type":      "Bearer",
			"expiresIn": 86400,
		})
	})
}

func RegisterAPIRoutes(r *gin.Engine, strategies storage.Strategies, jwt *auth.JWTManager, cfg *config.Config) {
	api := r.Group("/api/v1")
	api.Use(auth.JWTAuthMiddleware(jwt))

	api.GET("/providers/:provider/buckets/:bucketName/files", func(c *gin.Context) {
		provider := storage.CloudProvider(c.Param("provider"))
		bucket := c.Param("bucketName")
		strat, ok := strategies[provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider not enabled"})
			return
		}
		files, err := strat.ListFiles(bucket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, files)
	})

	api.GET("/providers/:provider/buckets/:bucketName/files/:fileName/download", func(c *gin.Context) {
		provider := storage.CloudProvider(c.Param("provider"))
		bucket := c.Param("bucketName")
		file := c.Param("fileName")
		strat, ok := strategies[provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider not enabled"})
			return
		}
		rc, err := strat.DownloadFile(bucket, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rc.Close()
		c.Header("Content-Disposition", "attachment; filename=\""+file+"\"")
		c.Header("Content-Type", "application/octet-stream")
		io.Copy(c.Writer, rc)
	})

	api.PUT("/providers/:provider/buckets/:bucketName/upload", func(c *gin.Context) {
		provider := storage.CloudProvider(c.Param("provider"))
		bucket := c.Param("bucketName")
		strat, ok := strategies[provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider not enabled"})
			return
		}
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
			return
		}
		defer file.Close()
		err = strat.UploadFile(bucket, fileHeader.Filename, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusOK)
	})

	api.GET("/providers/:provider/buckets/:bucketName/files/:fileName/presigned-url", func(c *gin.Context) {
		provider := storage.CloudProvider(c.Param("provider"))
		bucket := c.Param("bucketName")
		file := c.Param("fileName")
		strat, ok := strategies[provider]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider not enabled"})
			return
		}
		url, err := strat.PresignedURL(bucket, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"presignedUrl": url})
	})
}

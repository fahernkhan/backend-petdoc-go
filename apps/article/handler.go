package article

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateArticle godoc
// @Summary Create new article
// @Tags Articles
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param title formData string true "Article title"
// @Param description formData string true "Article content"
// @Param category formData string true "Article category"
// @Param image formData file true "Article image"
// @Success 201 {object} ArticleResponse
// @Failure 400 {object} map[string]string
// @Router /articles [post]
func (h *Handler) CreateArticle(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	article, err := h.service.CreateArticle(c.Request.Context(), req, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, article)
}

// GetArticle godoc
// @Summary Get article by ID
// @Tags Articles
// @Produce json
// @Param id path int true "Article ID"
// @Success 200 {object} ArticleResponse
// @Failure 404 {object} map[string]string
// @Router /articles/{id} [get]
func (h *Handler) GetArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.service.GetArticle(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

// GetAllArticles godoc
// @Summary Get all articles
// @Tags Articles
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} PaginationResponse
// @Failure 400 {object} map[string]string
// @Router /articles [get]
func (h *Handler) GetAllArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	articles, err := h.service.GetAllArticles(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, articles)
}

// UpdateArticle godoc
// @Summary Update article
// @Tags Articles
// @Accept multipart/form-data
// @Produce json
// UpdateArticle godoc
// @Summary Update article
// @Tags Articles
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Article ID"
// @Param title formData string false "Article title"
// @Param description formData string false "Article content"
// @Param category formData string false "Article category"
// @Param image formData file false "Article image"
// @Success 200 {object} ArticleResponse
// @Failure 400 {object} map[string]string
// @Router /articles/{id} [put]
func (h *Handler) UpdateArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt("user_id")
	article, err := h.service.UpdateArticle(c.Request.Context(), req, id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

// DeleteArticle godoc
// @Summary Delete article
// @Tags Articles
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Article ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /articles/{id} [delete]
func (h *Handler) DeleteArticle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.service.DeleteArticle(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

// SearchArticles godoc
// @Summary Search articles
// @Tags Articles
// @Produce json
// @Param query query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(10)
// @Success 200 {object} PaginationResponse
// @Failure 400 {object} map[string]string
// @Router /articles/search [get]
func (h *Handler) SearchArticles(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	articles, err := h.service.SearchArticles(c.Request.Context(), query, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, articles)
}

func handleError(c *gin.Context, err error) {
	switch err {
	case ErrArticleNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUnauthorizedAccess:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case ErrInvalidImageFormat:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrPaginationInvalid:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

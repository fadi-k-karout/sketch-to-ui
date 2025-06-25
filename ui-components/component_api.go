package uicomponents

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"sketch-to-ui-final-proj/ai"
	"sketch-to-ui-final-proj/auth"
	"sketch-to-ui-final-proj/sketch"

	"github.com/gin-gonic/gin"
)

// UIComponentHandler handles all UI component related operations
type UIComponentHandler struct {
	componentStore *UIComponentsStore
	sketchStore    *sketch.SketchStore
	aiProvider     *ai.OpenRouterProvider
}

// NewUIComponentHandler creates a new instance of UIComponentHandler
func NewUIComponentHandler(componentStore *UIComponentsStore, sketchStore *sketch.SketchStore, aiProvider *ai.OpenRouterProvider) *UIComponentHandler {
	return &UIComponentHandler{
		componentStore: componentStore,
		sketchStore:    sketchStore,
		aiProvider:     aiProvider,
	}
}

// CreateComponentRequest represents the request payload for creating a new component
type CreateComponentRequest struct {
	SketchID    string `form:"sketch_id" binding:"required"`
	UserPrompt  string `form:"user_prompt" binding:"omitempty"`
	Title       string `form:"title" binding:"max=20,omitempty"`
	Description string `form:"description" binding:="omitempty"`
	IsPublic    bool
}

// UpdateComponentRequest represents the request payload for updating a component
type UpdateComponentRequest struct {
	Title       string `form:"title,omitempty"`
	Type        string `form:"type,omitempty"`
	Code        string `form:"code,omitempty"`
	Description string `form:"description,omitempty"`
}

type PaginationQuery struct {
	Limit  int `form:"limit" binding:"max=20"`
	Offset int `form:"offset"`
}

// CreateComponent handles POST requests to create a new UI component from a sketch
func (h *UIComponentHandler) CreateComponent(c *gin.Context) {
	var req CreateComponentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Get user ID from context, set by auth middleware
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get the sketch from the sketch store
	sketch, _, err := h.sketchStore.GetSketch(req.SketchID)
	if err != nil || sketch == nil {
		slog.Error("Failed to get sketch", "sketch_id", req.SketchID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Sketch not found"})
		return
	}

	if sketch.ImageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sketch does not have a valid image"})
		return
	}

	// Generate UI code using the AI package
	userPrompt := req.UserPrompt
	if userPrompt == "" {
		userPrompt = "Analyze the following sketch image from the image url I sent and generate the corresponding UI component code (using  HTML and CSS in a style tag above the HTML code) in JSON format."
	}

	uiGenResp, err := ai.GenerateUICode(c.Request.Context(), userPrompt, "data:image/png;base64,"+sketch.ImageURL, h.aiProvider)
	if err != nil {
		slog.Error("Failed to generate UI code", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate UI components"})
		return
	}

	if len(uiGenResp.Components) == 0 {
		failureMsg := uiGenResp.FailureResponse
		if failureMsg == "" {
			failureMsg = "Failed to generate UI components from the provided sketch"
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": failureMsg})
		return
	}

	// NOTE: If your componentStore supports transactions, it would be best to wrap
	// the following loop in a transaction to ensure all components are created or none are.
	var createdComponents []UIComponent
	for _, dto := range uiGenResp.Components {
		component := UIComponent{
			UserID: userID, // Associate the component with the user
			Title:  dto.Title,
			Type:   dto.Type,
			Code:   dto.Code,
		}

		// Override title if provided in request and only one component is generated
		if req.Title != "" && len(uiGenResp.Components) == 1 {
			component.Title = req.Title
		}

		if err := h.componentStore.CreateComponent(&component); err != nil {
			slog.Error("Failed to create component", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save component"})
			return // Exit on first error
		}

		createdComponents = append(createdComponents, component)
	}

	location := map[string]interface{}{
		"path":   "/components/dashboard",
		"target": "#content",
	}
	locationJSON, err := json.Marshal(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal HX-Location"})
		return
	}
	c.Header("HX-Location", string(locationJSON))
	c.Status(http.StatusOK)

}

// UpdateComponent handles PUT requests to update an existing UI component
func (h *UIComponentHandler) UpdateComponent(c *gin.Context) {
	// Get component ID from URL parameter
	componentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	var req UpdateComponentRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
		return
	}

	// Get existing component
	existingComponent, err := h.componentStore.GetComponentByID(componentID)
	if err != nil {
		slog.Error("Failed to get component", "component_id", componentID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	// SECURITY: Check if the user owns this component
	if existingComponent.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this component"})
		return
	}

	// Update fields if provided in the request
	if req.Title != "" {
		existingComponent.Title = req.Title
	}
	if req.Type != "" {
		existingComponent.Type = req.Type
	}
	if req.Code != "" {
		existingComponent.Code = req.Code
	}

	// Update in database
	if err = h.componentStore.UpdateComponent(existingComponent.ID, existingComponent); err != nil {
		slog.Error("Failed to update component", "component_id", componentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component"})
		return
	}

	location := map[string]interface{}{
		"path":   "/components/dashboard",
		"target": "#content",
	}
	locationJSON, err := json.Marshal(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal HX-Location"})
		return
	}

	c.Header("HX-Location", string(locationJSON))
	c.Status(http.StatusOK)
}

// ArchiveComponent handles DELETE requests to archive a UI component
func (h *UIComponentHandler) ArchiveComponent(c *gin.Context) {
	// Get component ID from URL parameter
	componentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}

	// Get user ID from context
	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized access"})
		return
	}

	// Get existing component to verify ownership
	existingComponent, err := h.componentStore.GetComponentByID(componentID)
	if err != nil {
		slog.Error("Failed to get component", "component_id", componentID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	// SECURITY: Check if the user owns this component
	if existingComponent.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to archive this component"})
		return
	}

	// Archive the component
	if err = h.componentStore.ArchiveComponent(componentID); err != nil {
		slog.Error("Failed to archive component", "component_id", componentID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive component"})
		return
	}

	c.Status(http.StatusOK)
}

// RenderComponentsMain handles GET requests to render the main components template
func (h *UIComponentHandler) RenderComponents(c *gin.Context) {
	var pagination PaginationQuery
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	userID, exists := auth.GetUserIDFromContext(c)
	slog.Debug("userID in RenderComponents Handler: ", "userID", userID)
	if !exists {
		c.HTML(http.StatusUnauthorized, "error.html", gin.H{"error": "Unauthorized access"})
		return
	}

	// Set default pagination if not provided
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}

	components, total, err := h.componentStore.GetComponentsByUserPaginated(userID, pagination.Limit, pagination.Offset)
	if err != nil {
		slog.Error("Failed to get components", "error", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load components"})
		return
	}

	nextOffset := pagination.Offset + len(components)
	remaining := total - nextOffset

	// Render the main components template
	c.HTML(http.StatusOK, "components-rows.html", gin.H{
		"Components": components,
		"NextOffset": nextOffset,
		"Remaining":  remaining,
	})
}

func (h *UIComponentHandler) RenderComponentsMain(c *gin.Context) {

	c.HTML(http.StatusOK, "component-main.html", gin.H{})

}

func (h *UIComponentHandler) RenderComponentsEdit(c *gin.Context) {

	userID, exists := auth.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Access"})
		return
	}
	componentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}
	component, err := h.componentStore.GetComponentByID(componentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Component not found"})
		slog.Error("Error Loading the component from the store", "error", err)
		return
	}

	if component.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to access this resource"})
		return
	}
	slog.Debug("Component is:", "component", component)

	c.HTML(http.StatusOK, "edit-view.html", gin.H{
		"Component": component,
	})
}

func (h *UIComponentHandler) RenderComponentsCreate(c *gin.Context) {

	c.HTML(http.StatusOK, "create-view.html", gin.H{})
}

// RegisterRoutes registers all component-related routes with the Gin router
func (h *UIComponentHandler) RegisterRoutes(router *gin.Engine) {
	componentGroup := router.Group("/components")
	componentGroup.Use(auth.AuthRequiredMiddleware())

	componentGroup.POST("/", h.CreateComponent)
	componentGroup.PUT("/:id", h.UpdateComponent) // Changed to use ID in path
	componentGroup.DELETE("/:id", h.ArchiveComponent)
	componentGroup.GET("/", h.RenderComponents)
	componentGroup.GET("/dashboard", h.RenderComponentsMain)

	componentGroup.GET("/create", h.RenderComponentsCreate)
	componentGroup.GET("/:id/edit", h.RenderComponentsEdit)
}

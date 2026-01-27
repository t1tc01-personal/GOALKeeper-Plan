package router

import (
	blockController "goalkeeper-plan/internal/block/controller"
	pageController "goalkeeper-plan/internal/page/controller"
	"goalkeeper-plan/internal/workspace/controller"

	"github.com/gin-gonic/gin"
)

// NewRouter registers workspace, page, block, and sharing-related routes under the given base group.
func NewRouter(baseRouter interface{}, workspaceController controller.WorkspaceController, pageCtl pageController.PageController, blockCtl blockController.BlockController, sharingCtl controller.SharingController) {
	group, ok := baseRouter.(*gin.RouterGroup)
	if !ok {
		return
	}

	// Workspace routes (protected - require auth)
	workspaces := group.Group("/notion/workspaces")
	{
		workspaces.GET("", workspaceController.ListWorkspaces)
		workspaces.POST("", workspaceController.CreateWorkspace)
		workspaces.GET("/:id", workspaceController.GetWorkspace)
		workspaces.PUT("/:id", workspaceController.UpdateWorkspace)
		workspaces.DELETE("/:id", workspaceController.DeleteWorkspace)
	}

	// Page routes
	pages := group.Group("/notion/pages")
	{
		pages.GET("", pageCtl.ListPages)
		pages.POST("", pageCtl.CreatePage)
		pages.GET("/:id", pageCtl.GetPage)
		pages.PUT("/:id", pageCtl.UpdatePage)
		pages.DELETE("/:id", pageCtl.DeletePage)
		pages.GET("/:id/hierarchy", pageCtl.GetHierarchy)
		// Sharing/Collaboration routes
		pages.GET("/:id/collaborators", sharingCtl.ListCollaborators)
		pages.POST("/:id/share", sharingCtl.GrantAccess)
		pages.DELETE("/:id/share/:user_id", sharingCtl.RevokeAccess)
	}

	// Block routes
	blocks := group.Group("/notion/blocks")
	{
		blocks.GET("", blockCtl.ListBlocks)
		blocks.POST("", blockCtl.CreateBlock)
		blocks.GET("/:id", blockCtl.GetBlock)
		blocks.PUT("/:id", blockCtl.UpdateBlock)
		blocks.DELETE("/:id", blockCtl.DeleteBlock)
		blocks.POST("/reorder", blockCtl.ReorderBlocks)
	}
}

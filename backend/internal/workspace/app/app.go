package app

import (
	"goalkeeper-plan/config"
	blockController "goalkeeper-plan/internal/block/controller"
	blockRepository "goalkeeper-plan/internal/block/repository"
	blockService "goalkeeper-plan/internal/block/service"
	"goalkeeper-plan/internal/logger"
	pageController "goalkeeper-plan/internal/page/controller"
	pageRepository "goalkeeper-plan/internal/page/repository"
	pageService "goalkeeper-plan/internal/page/service"
	"goalkeeper-plan/internal/workspace/controller"
	"goalkeeper-plan/internal/workspace/repository"
	"goalkeeper-plan/internal/workspace/router"
	"goalkeeper-plan/internal/workspace/service"

	"gorm.io/gorm"
)

// NewApplication wires the workspace module following the internal/*
// pattern (repository -> service -> controller -> router).
func NewApplication(db *gorm.DB, baseRouter interface{}, configs config.Configurations, log logger.Logger) {
	workspaceRepo := repository.NewWorkspaceRepository(db)
	pageRepo := pageRepository.NewPageRepository(db)
	blockRepo := blockRepository.NewBlockRepository(db)
	sharingPermRepo := repository.NewSharePermissionRepository(db)

	workspaceService, err := service.NewWorkspaceService(
		service.WithWorkspaceRepository(workspaceRepo),
		service.WithLogger(log),
	)
	if err != nil {
		log.Fatal("failed to initialize workspace service")
		return
	}

	pageSvc, err := pageService.NewPageService(
		pageService.WithPageRepository(pageRepo),
		pageService.WithPageLogger(log),
	)
	if err != nil {
		log.Fatal("failed to initialize page service")
		return
	}

	blockSvc, err := blockService.NewBlockService(
		blockService.WithBlockRepository(blockRepo),
		blockService.WithBlockLogger(log),
	)
	if err != nil {
		log.Fatal("failed to initialize block service")
		return
	}

	sharingSvc, err := service.NewSharingService(
		service.WithSharingRepository(sharingPermRepo),
		service.WithSharingLogger(log),
	)
	if err != nil {
		log.Fatal("failed to initialize sharing service")
		return
	}

	workspaceController := controller.NewWorkspaceController(workspaceService, log)
	pageCtl := pageController.NewPageController(pageSvc, log)
	blockCtl := blockController.NewBlockController(blockSvc, pageSvc, log)
	sharingCtl := controller.NewSharingController(sharingSvc)

	router.NewRouter(baseRouter, workspaceController, pageCtl, blockCtl, sharingCtl)
}

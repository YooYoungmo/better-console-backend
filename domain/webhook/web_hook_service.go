package webhook

import (
	"better-console-backend/domain"
	"better-console-backend/dtos"
	"context"
)

type WebHookService struct {
}

func (WebHookService) CreateWebHook(ctx context.Context, webHookInformation dtos.WebHookInformation) error {
	repository := webHookRepository{}
	lastEntity, err := repository.FindLast(ctx)
	var nextId uint
	if err != nil {
		if err == domain.ErrNotFound {
			nextId = 1
		} else {
			return err
		}
	} else {
		nextId = lastEntity.nextId()
	}

	entity, err := NewWebHookEntity(nextId, webHookInformation)
	if err != nil {
		return err
	}

	return repository.Create(ctx, &entity)
}

func (WebHookService) GetWebHooks(ctx context.Context, pageable dtos.Pageable) ([]WebHookEntity, int64, error) {
	return webHookRepository{}.FindAll(ctx, pageable)
}

func (WebHookService) DeleteWebHook(ctx context.Context, webHookId uint) error {
	repository := webHookRepository{}

	entity, err := repository.FindById(ctx, webHookId)
	if err != nil {
		return err
	}

	return repository.Delete(ctx, entity)
}

func (WebHookService) GetWebHook(ctx context.Context, webHookId uint) (WebHookEntity, error) {
	return webHookRepository{}.FindById(ctx, webHookId)
}

func (WebHookService) UpdateWebHook(ctx context.Context, webHookId uint, webHookInformation dtos.WebHookInformation) error {
	repository := webHookRepository{}

	entity, err := repository.FindById(ctx, webHookId)
	if err != nil {
		return err
	}

	entity.Update(webHookInformation)

	return repository.Save(ctx, entity)
}

func (WebHookService) NoteMessage(ctx context.Context, webHookId uint, message dtos.WebHookMessage) error {
	repository := webHookRepository{}
	entity, err := repository.FindById(ctx, webHookId)
	if err != nil {
		return err
	}

	entity.AddMessage(message)

	err = repository.Save(ctx, entity)
	if err != nil {
		return err
	}

	message.Title = entity.Name

	return entity.NoteMessage(message)
}

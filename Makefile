.PHONY: deploy-preview deploy-production

WEBHOOK_PATH := /api/tg/webhook
TGVERCEL := go run github.com/harnyk/tgvercel


deploy-preview:
	$(TGVERCEL) hook $$(vercel) $(WEBHOOK_PATH)

deploy-production:
	$(TGVERCEL) hook $$(vercel --prod) $(WEBHOOK_PATH)

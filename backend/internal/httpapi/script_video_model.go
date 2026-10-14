package httpapi

import "qiantie/backend/internal/batchfactoryv11"

type scriptVideoModel struct {
	Provider               string
	AcceptsReferenceImages bool
	PublicTaskPrefix       string
}

func scriptVideoModelSpec(modelKey string) (scriptVideoModel, bool) {
	switch modelKey {
	case "yd2-mini-video":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderPersonalAPI, AcceptsReferenceImages: true}, true
	case "minimax-h3-video":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderAutoDLH3, AcceptsReferenceImages: true, PublicTaskPrefix: "h3:"}, true
	case "seedance-2-0-official":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderYFAISeedance, AcceptsReferenceImages: true, PublicTaskPrefix: "yfai:"}, true
	case "local-doubao-executor-video":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderDoubaoLocal, PublicTaskPrefix: localScriptVideoTaskPrefix}, true
	default:
		return scriptVideoModel{}, false
	}
}

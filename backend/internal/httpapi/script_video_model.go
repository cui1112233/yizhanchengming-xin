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
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderPersonalAPI, AcceptsReferenceImages: true, PublicTaskPrefix: "yd2:"}, true
	case "minimax-h3-video":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderAutoDLH3, AcceptsReferenceImages: true, PublicTaskPrefix: "h3:"}, true
	case "seedance-2-0-official":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderYFAISeedance, AcceptsReferenceImages: true, PublicTaskPrefix: "yfai:"}, true
	case "fanke-video-api":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderYFAISeedance, AcceptsReferenceImages: true, PublicTaskPrefix: "yfai:fanke-video-api:"}, true
	case "local-doubao-executor-video":
		return scriptVideoModel{Provider: batchfactoryv11.VideoProviderDoubaoLocal, AcceptsReferenceImages: true, PublicTaskPrefix: localScriptVideoTaskPrefix}, true
	default:
		return scriptVideoModel{}, false
	}
}

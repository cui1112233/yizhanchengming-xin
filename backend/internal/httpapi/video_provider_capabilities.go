package httpapi

import "qiantie/backend/internal/batchfactoryv11"

// VideoProviderCapability mirrors one entry of the shared front-end contract
// file. The JSON file is the single source consumed by the settings drawer and
// the admin model catalog; this Go mirror exists so the backend can reject
// unsupported providers before the request reaches the registry.
type VideoProviderCapability struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	DefaultModel string `json:"defaultModel"`
	CatalogID    string `json:"catalogId"`
	AdminAdapter string `json:"adminAdapter"`
	AdminLabel   string `json:"adminLabel"`
}

// videoProviderCapabilityContract mirrors
// frontend/src/user/pages/batch-factory-v11/videoProviderCapabilities.json.
// Keep both sides in sync; TestVideoProviderCapabilityContract guards it.
var videoProviderCapabilityContract = []VideoProviderCapability{
	{ID: batchfactoryv11.VideoProviderPersonalAPI, Label: "API 配置模型", DefaultModel: batchfactoryv11.DefaultPersonalVideoModel, CatalogID: "yd2-mini-video", AdminAdapter: "yd_video", AdminLabel: "YD2.0 Mini 图生视频"},
	{ID: batchfactoryv11.VideoProviderDoubaoLocal, Label: "本地执行器模型", DefaultModel: "doubao-seedance", AdminAdapter: "local_executor_video", AdminLabel: "本地豆包执行器"},
	{ID: batchfactoryv11.VideoProviderAutoDLH3, Label: "API 配置模型（AutoDL H3）", DefaultModel: batchfactoryv11.AutoDLH3Model, CatalogID: "minimax-h3-video", AdminAdapter: "yd_video", AdminLabel: "YD2.0 Mini 图生视频"},
	{ID: batchfactoryv11.VideoProviderYFAISeedance, Label: "梵客 Seedance（yfai）", DefaultModel: "seedance-2-0-official", CatalogID: "seedance-2-0-official", AdminAdapter: "yfai_seedance", AdminLabel: "梵客 Seedance 图生视频"},
}

// VideoProviderCapabilities returns the backend-side view of the contract.
func VideoProviderCapabilities() []VideoProviderCapability {
	return append([]VideoProviderCapability(nil), videoProviderCapabilityContract...)
}

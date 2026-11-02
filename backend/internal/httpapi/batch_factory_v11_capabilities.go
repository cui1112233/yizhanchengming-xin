package httpapi

import "net/http"

type Capability struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

func CapabilitiesForSlice(slice int) map[string]Capability {
	return capabilitiesForRuntime(slice, slice >= 4, slice >= 5)
}

func capabilitiesForRuntime(slice int, productionAvailable bool, mergeAvailable bool, externalAvailability ...bool) map[string]Capability {
	caps := map[string]Capability{
		"batch.read":          {Reason: "V11 settings slice not released"},
		"batch.create":        {Reason: "V11 settings slice not released"},
		"settings.edit":       {Reason: "V11 settings slice not released"},
		"snapshot.read":       {Reason: "V11 settings slice not released"},
		"override.edit":       {Reason: "V11 settings slice not released"},
		"director.run":        {Reason: "Director slice not released"},
		"hook.review":         {Reason: "Director slice not released"},
		"working-front.viral": {Reason: "Director slice not released"},
		"compiler.preview":    {Reason: "Final prompt compiler slice not released"},
		"production.submit":   {Reason: "Production slice not released"},
		"production.cancel":   {Reason: "Production slice not released"},
		"merge.run":           {Reason: "Merge slice not released"},
		"publish.121":         {Reason: "121 视频发布未启用：尚未验证 MP4 提交接口、回执与回读"},
		"publish.yadi":        {Reason: "Yadi is not enabled"},
	}
	if slice >= 1 {
		for _, key := range []string{"batch.read", "batch.create", "settings.edit", "snapshot.read", "override.edit"} {
			caps[key] = Capability{Available: true}
		}
	}
	if slice >= 2 {
		caps["director.run"] = Capability{Available: true}
		caps["hook.review"] = Capability{Available: true}
		caps["working-front.viral"] = Capability{Available: true}
	}
	if slice >= 3 {
		caps["compiler.preview"] = Capability{Available: true}
	}
	if slice >= 4 && productionAvailable {
		caps["production.submit"] = Capability{Available: true}
		caps["production.cancel"] = Capability{Available: true}
	}
	if slice >= 5 && mergeAvailable {
		caps["merge.run"] = Capability{Available: true}
	}
	// The Go-owned 121 publisher is registered alongside the merge flow. It
	// can run as soon as slice 5 has enabled final-video production; account,
	// source metadata, and target-site readiness are validated by the actual
	// upload request instead of being hidden behind a stale release gate.
	if slice >= 5 && len(externalAvailability) > 0 && externalAvailability[0] {
		caps["publish.121"] = Capability{Available: true}
	}
	if slice >= 6 && len(externalAvailability) > 1 && externalAvailability[1] {
		caps["publish.yadi"] = Capability{Available: true}
	}
	return caps
}

func capabilityHandler(slice int) http.HandlerFunc {
	return capabilityHandlerForRuntime(slice, slice >= 4, slice >= 5)
}

func capabilityHandlerForRuntime(slice int, productionAvailable bool, mergeAvailability ...bool) http.HandlerFunc {
	mergeAvailable := slice >= 5
	var publish121, publishYadi bool
	if len(mergeAvailability) > 0 {
		mergeAvailable = mergeAvailability[0]
	}
	if len(mergeAvailability) > 1 {
		publish121 = mergeAvailability[1]
	}
	if len(mergeAvailability) > 2 {
		publishYadi = mergeAvailability[2]
	}
	return func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusOK, capabilitiesForRuntime(slice, productionAvailable, mergeAvailable, publish121, publishYadi))
	}
}

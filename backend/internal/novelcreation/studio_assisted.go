package novelcreation

import "context"

// Routes are server-owned. Never navigate to a model-supplied URL or create a
// book as a side effect of opening an intake form.
func studioAssistedRoute(kind string) string {
	switch kind {
	case "style_imitation":
		return "imitation"
	case "continuation_import":
		return "chapters"
	default:
		return ""
	}
}

func (s *Service) confirmStudioAssisted(ctx context.Context, owner, id, actionID string) (out StudioTransition, err error) {
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		r, e := readStudio(ctx, tx, owner, id, true)
		if e != nil {
			return e
		}
		if r.Action == nil || r.Action.ID != actionID {
			return conflict()
		}
		if r.Action.Status == "confirmed" && r.Transition != nil {
			out = *r.Transition
			return nil
		}
		route := studioAssistedRoute(r.Action.Type)
		if r.Action.Status != "pending" || route == "" {
			return conflict()
		}
		r.Action.Status = "confirmed"
		out = StudioTransition{Session: r.Session, AssistedAction: r.Action.Type, TargetRoute: route}
		r.Transition = &out
		return writeStudio(ctx, tx, owner, r, false)
	})
	return
}

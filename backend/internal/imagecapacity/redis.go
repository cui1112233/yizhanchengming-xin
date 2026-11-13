package imagecapacity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand/v2"
	"sync"
	"time"
)

// Every state transition, including expiry and hierarchical fair selection, occurs
// in one Redis operation. A lost worker leaves its reservation for the provider's
// safe horizon; heartbeat expiry never immediately creates replacement capacity.
const stateScript = `
local action=ARGV[1]; local a=cjson.decode(ARGV[2])
if action=='acquire' or action=='snapshot' then
 if redis.call('EXISTS',KEYS[2])==1 or redis.call('EXISTS',KEYS[4])==1 or (redis.call('GET',KEYS[3]) or '0')~=a.epoch then return cjson.encode({granted=-1}) end
end
local raw=redis.call('GET',KEYS[1])
local s=raw and cjson.decode(raw) or {leases={},waiting={},profiles={},owners={},books={},sequence=0}
s.models=s.models or {}
local tm=redis.call('TIME'); local now=tonumber(tm[1])*1000+math.floor(tonumber(tm[2])/1000)
local profile=a.profile
for id,v in pairs(s.leases) do if v.deadline<=now then s.leases[id]=nil end end
for id,v in pairs(s.waiting) do if v.deadline<=now then s.waiting[id]=nil end end
local p=s.profiles[profile]
if not p then p={limit=a.initial,hard=a.hard,cooling=0,successes=0,lastBad=0,stable=0,revision=a.revision};s.profiles[profile]=p end
if action=='acquire' or action=='snapshot' then
 if p.revision~=a.revision then p.limit=a.initial;p.successes=0;p.stable=0;p.revision=a.revision end
 p.hard=a.hard;p.limit=math.min(p.limit,p.hard);s.models[a.model]=a.modeltotal
end
local function counts()
 local running=0;local model=0;local waiting=0;local uncertain=0;local modeltotal=0
 for _,v in pairs(s.leases) do if not v.model or v.model==a.model then modeltotal=modeltotal+(v.units or 1) end;running=running+(v.units or 1);if v.profile==profile then model=model+(v.units or 1) end;if v.uncertain then uncertain=uncertain+(v.units or 1) end end
 for _,v in pairs(s.waiting) do waiting=waiting+1 end
 return running,model,waiting,uncertain,modeltotal
end
local byProfile={};local byModel={};local legacyUnits=0
for _,v in pairs(s.leases) do
 byProfile[v.profile]=(byProfile[v.profile] or 0)+(v.units or 1)
 if v.model then byModel[v.model]=(byModel[v.model] or 0)+(v.units or 1) else legacyUnits=legacyUnits+(v.units or 1) end
end
local granted=0
if action=='acquire' then
 if s.leases[a.token] then granted=1 elseif a.units>a.account or a.units>p.hard or a.units>a.modeltotal then s.waiting[a.token]=nil;granted=-2 else
  if not s.waiting[a.token] then s.sequence=s.sequence+1;s.waiting[a.token]={owner=a.owner,book=a.book,profile=profile,model=a.model,units=a.units,seq=s.sequence,deadline=now+30000} else s.waiting[a.token].deadline=now+30000;s.waiting[a.token].model=a.model end
  local running,model,_,_,modeltotal=counts()
  if running+a.units<=a.account and model+a.units<=p.limit and modeltotal+a.units<=a.modeltotal and now>=p.cooling then
   local best=nil;local bo=nil;local bb=nil;local bs=nil
   for id,v in pairs(s.waiting) do
    local vp=s.profiles[v.profile];local n=byProfile[v.profile] or 0;local total=running
    if v.model then total=(byModel[v.model] or 0)+legacyUnits end
    if vp and n+(v.units or 1)<=vp.limit and running+(v.units or 1)<=a.account and total+(v.units or 1)<=(s.models[v.model or ""] or 30) and now>=vp.cooling then
     local o=s.owners[v.owner] or 0;local b=s.books[v.owner..':'..v.book] or 0
     if not best or o<bo or (o==bo and (b<bb or (b==bb and v.seq<bs))) then best=id;bo=o;bb=b;bs=v.seq end
    end
   end
   if best==a.token then
    local w=s.waiting[a.token];s.waiting[a.token]=nil;s.sequence=s.sequence+1
    s.owners[w.owner]=s.sequence;s.books[w.owner..':'..w.book]=s.sequence
    s.leases[a.token]={profile=profile,model=a.model,units=a.units,deadline=now+a.horizon};granted=1
   end
  end
 end
elseif action=='cancel' then s.waiting[a.token]=nil
elseif action=='renew' then
 if s.leases[a.token] then if not s.leases[a.token].uncertain then s.leases[a.token].deadline=now+a.horizon end;granted=1 end
elseif action=='finish' then
 if s.leases[a.token] then
  if a.unknown then s.leases[a.token].deadline=now+a.horizon;s.leases[a.token].uncertain=true else s.leases[a.token]=nil end
  if a.limited then
   p.limit=math.min(p.hard,math.max(a.units,math.floor(p.limit/2)));p.cooling=math.max(p.cooling,now+a.retry);p.successes=0;p.stable=0;p.lastBad=now
  elseif a.transient then
   p.failures=(p.failures or 0)+1;p.successes=0;p.stable=0
   if p.failures>=3 then p.limit=math.min(p.hard,math.max(a.units,math.floor(p.limit/2)));p.cooling=math.max(p.cooling,now+10000);p.failures=0;p.lastBad=now end
  elseif a.success and not a.unknown then
   p.failures=0
   if now>=p.cooling then
    if p.stable==0 then p.stable=now end;p.successes=p.successes+1
    if p.successes>=20 and now-p.stable>=30000 and now-p.lastBad>=30000 then p.limit=math.min(p.hard,p.limit+1);p.successes=0;p.stable=now end
   end
  end
 end
end
local running,model,waiting,uncertain,modeltotal=counts()
redis.call('SET',KEYS[1],cjson.encode(s),'PX',7200000)
return cjson.encode({granted=granted,running=running,model=model,waiting=waiting,uncertain=uncertain,modeltotal=modeltotal,effective=p.limit,cooling=p.cooling})
`

type Controller struct {
	repo      Repository
	do        func(context.Context, ...string) (any, error)
	prefix    string
	heartbeat time.Duration
	horizon   time.Duration
	cacheTTL  time.Duration
	cacheMu   sync.Mutex
	cache     map[string]cachedPolicy
}
type cachedPolicy struct {
	policy  Policy
	expires time.Time
}

// A crashed writer's dirty marker survives its 60s lock. Recovering readers bump
// the epoch before reading MySQL; all pre-crash policy reads are thereby fenced.
const configReadScript = `
if redis.call('EXISTS',KEYS[1])==0 and redis.call('EXISTS',KEYS[3])==1 then
 redis.call('INCR',KEYS[2]);redis.call('DEL',KEYS[3])
end
return redis.call('GET',KEYS[2]) or '0'
`
const configLockScript = `
if redis.call('EXISTS',KEYS[1])==1 then return 0 end
if redis.call('EXISTS',KEYS[3])==1 then redis.call('INCR',KEYS[2]);redis.call('DEL',KEYS[3]) end
redis.call('INCR',KEYS[2]);redis.call('SET',KEYS[1],ARGV[1],'PX',60000);redis.call('SET',KEYS[3],ARGV[1]);return 1
`
const configUnlockScript = `
if redis.call('GET',KEYS[1])~=ARGV[1] then return 0 end
redis.call('INCR',KEYS[2]);redis.call('DEL',KEYS[1],KEYS[3]);return 1
`

func waitPoll(ctx context.Context) error {
	t := time.NewTimer(time.Duration(250+mathrand.IntN(100)) * time.Millisecond)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// SavePolicy is the sole safe runtime mutation entry point. Policy writes are
// bounded to 15s, well within the barrier's 60s crash-recovery quarantine.
func (c *Controller) SavePolicy(ctx context.Context, r Request, p Policy) (retErr error) {
	if c == nil || c.do == nil || c.repo == nil {
		return fmt.Errorf("image capacity unavailable")
	}
	if p.AccountKey != AccountKey(r) || p.ProfileKey != ProfileKey(r) || p.AccountLimit < 1 || p.ModelLimit < 1 || p.InitialLimit < 1 || p.InitialLimit > p.ModelLimit {
		return fmt.Errorf("invalid image capacity policy")
	}
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return e
	}
	token := hex.EncodeToString(b)
	keys := []string{c.prefix + ":config:lock", c.prefix + ":config:epoch", c.prefix + ":config:dirty"}
	for {
		v, e := c.do(ctx, "EVAL", configLockScript, "3", keys[0], keys[1], keys[2], token)
		if e != nil {
			return e
		}
		if fmt.Sprint(v) == "1" {
			break
		}
		if e = waitPoll(ctx); e != nil {
			return e
		}
	}
	quarantine := false
	defer func() {
		// A commit acknowledgement can be lost even when MySQL committed. Keep
		// the crash quarantine on uncertain writes instead of serving stale reads.
		if quarantine {
			return
		}
		x, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		v, e := c.do(x, "EVAL", configUnlockScript, "3", keys[0], keys[1], keys[2], token)
		if e != nil || fmt.Sprint(v) != "1" {
			if retErr == nil {
				retErr = fmt.Errorf("image capacity configuration barrier lost: %v", e)
			}
		}
	}()
	old, e := c.repo.Resolve(ctx, r)
	if e != nil {
		return e
	}
	old = normalize(old, r)
	if old.GroupKey != p.GroupKey {
		for _, pool := range []Policy{old, p} {
			v, e := c.transition(ctx, "inspect", pool, r, "", Outcome{})
			if e != nil {
				return e
			}
			if v.Running > 0 || v.Waiting > 0 {
				return fmt.Errorf("cannot change account group while old or target pool has active or waiting images")
			}
		}
	}
	dbctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	retErr = c.repo.Save(dbctx, p)
	quarantine = retErr != nil
	return retErr
}

func NewController(repo Repository, do func(context.Context, ...string) (any, error), prefix string) *Controller {
	if prefix == "" {
		prefix = "qiantie:imagecapacity"
	}
	return &Controller{repo: repo, do: do, prefix: prefix, heartbeat: 10 * time.Second, horizon: 30 * time.Minute, cacheTTL: time.Second, cache: make(map[string]cachedPolicy)}
}
func (c *Controller) Resolve(ctx context.Context, r Request) (Policy, error) {
	if c == nil || c.repo == nil || c.do == nil {
		return Policy{}, fmt.Errorf("image capacity policy unavailable")
	}
	v, e := c.do(ctx, "EVAL", configReadScript, "3", c.prefix+":config:lock", c.prefix+":config:epoch", c.prefix+":config:dirty")
	if e != nil {
		return Policy{}, fmt.Errorf("image capacity Redis: %w", e)
	}
	cacheKey := AccountKey(r) + ":" + ProfileKey(r) + ":" + fmt.Sprint(v)
	if c.cacheTTL > 0 {
		c.cacheMu.Lock()
		entry, ok := c.cache[cacheKey]
		c.cacheMu.Unlock()
		if ok && time.Now().Before(entry.expires) {
			return entry.policy, nil
		}
	}
	p, e := c.repo.Resolve(ctx, r)
	p.configEpoch = fmt.Sprint(v)
	p = normalize(p, r)
	if e == nil && c.cacheTTL > 0 {
		c.cacheMu.Lock()
		if c.cache == nil {
			c.cache = make(map[string]cachedPolicy)
		}
		if len(c.cache) >= 1024 {
			for k, entry := range c.cache {
				if time.Now().After(entry.expires) {
					delete(c.cache, k)
				}
			}
			if len(c.cache) >= 1024 {
				for k := range c.cache {
					delete(c.cache, k)
					break
				}
			}
		}
		c.cache[cacheKey] = cachedPolicy{policy: p, expires: time.Now().Add(c.cacheTTL)}
		c.cacheMu.Unlock()
	}
	return p, e
}

type Status struct {
	Policy            Policy    `json:"policy"`
	Running           int       `json:"running"`
	Uncertain         int       `json:"uncertain"`
	ModelTotalRunning int       `json:"modelTotalRunning"`
	ModelRunning      int       `json:"modelRunning"`
	Waiting           int       `json:"waiting"`
	EffectiveLimit    int       `json:"effectiveLimit"`
	CoolingUntil      time.Time `json:"coolingUntil"`
}
type Outcome struct {
	Success, Limited, Transient, Unknown bool
	RetryAfter, Duration                 time.Duration
}
type stateResult struct {
	Granted, Running, Model, Waiting, Effective, Uncertain, ModelTotal int
	Cooling                                                            int64
}

func (c *Controller) transition(ctx context.Context, action string, p Policy, r Request, token string, o Outcome) (stateResult, error) {
	var result stateResult
	if c == nil || c.do == nil {
		return result, fmt.Errorf("image capacity Redis unavailable")
	}
	group := p.AccountKey
	if p.GroupKey != "" {
		group = "group:" + p.GroupKey
	}
	retry := o.RetryAfter
	if retry <= 0 {
		retry = 10 * time.Second
	}
	body, _ := json.Marshal(map[string]any{"model": hash(modelIdentity(r.Model)), "modeltotal": p.ModelTotalLimit, "units": imageCount(r.Count), "epoch": p.configEpoch, "profile": p.ProfileKey, "initial": p.InitialLimit, "hard": p.ModelLimit, "revision": hash([]any{p.ModelLimit, p.InitialLimit, p.Verified}), "account": p.AccountLimit, "token": token, "owner": hash(r.Owner), "book": hash(r.Book), "horizon": c.horizon.Milliseconds(), "limited": o.Limited, "unknown": o.Unknown, "transient": o.Transient, "success": o.Success, "retry": retry.Milliseconds()})
	v, e := c.do(ctx, "EVAL", stateScript, "4", c.prefix+":"+hash(group), c.prefix+":config:lock", c.prefix+":config:epoch", c.prefix+":config:dirty", action, string(body))
	if e != nil {
		return result, fmt.Errorf("image capacity Redis: %w", e)
	}
	switch v := v.(type) {
	case string:
		e = json.Unmarshal([]byte(v), &result)
	case []byte:
		e = json.Unmarshal(v, &result)
	default:
		e = fmt.Errorf("invalid Redis response")
	}
	return result, e
}
func (c *Controller) Snapshot(ctx context.Context, r Request) (Status, error) {
	for {
		p, e := c.Resolve(ctx, r)
		if e != nil {
			return Status{}, e
		}
		v, e := c.transition(ctx, "snapshot", p, r, "", Outcome{})
		if e == nil && v.Granted == -1 {
			if e = waitPoll(ctx); e != nil {
				return Status{}, e
			}
			continue
		}
		st := Status{Policy: p, Running: v.Running, Uncertain: v.Uncertain, ModelRunning: v.Model, ModelTotalRunning: v.ModelTotal, Waiting: v.Waiting, EffectiveLimit: v.Effective}
		if v.Cooling > 0 {
			st.CoolingUntil = time.UnixMilli(v.Cooling)
		}
		return st, e
	}
}
func (c *Controller) Acquire(ctx context.Context, r Request) (*Lease, error) {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return nil, e
	}
	token := hex.EncodeToString(b)
	p, e := c.Resolve(ctx, r)
	if e != nil {
		return nil, e
	}
	cleanup := func() {
		x, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		c.transition(x, "cancel", p, r, token, Outcome{})
	}
	for {
		if e = ctx.Err(); e != nil {
			cleanup()
			return nil, e
		}
		p, e = c.Resolve(ctx, r)
		if e != nil {
			cleanup()
			return nil, e
		}
		v, e := c.transition(ctx, "acquire", p, r, token, Outcome{})
		if e != nil {
			cleanup()
			return nil, e
		}
		if v.Granted == -2 {
			cleanup()
			return nil, fmt.Errorf("image count %d exceeds account/model hard capacity", imageCount(r.Count))
		}
		if v.Granted == 1 {
			lc, cancel := context.WithCancel(ctx)
			l := &Lease{c: c, p: p, r: r, token: token, ctx: lc, cancel: cancel, stop: make(chan struct{})}
			go l.renew()
			return l, nil
		}
		if e = waitPoll(ctx); e != nil {
			cleanup()
			return nil, e
		}
	}
}

type Lease struct {
	c      *Controller
	p      Policy
	r      Request
	token  string
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	stop   chan struct{}
}

func (l *Lease) Context() context.Context { return l.ctx }
func (l *Lease) renew() {
	t := time.NewTicker(l.c.heartbeat)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-l.ctx.Done():
			return
		case <-t.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			v, e := l.c.transition(ctx, "renew", l.p, l.r, l.token, Outcome{})
			cancel()
			if e != nil || v.Granted != 1 {
				l.cancel()
				return
			}
		}
	}
}
func (l *Lease) Finish(o Outcome) {
	if l == nil {
		return
	}
	l.once.Do(func() {
		close(l.stop)
		l.cancel()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		l.c.transition(ctx, "finish", l.p, l.r, l.token, o)
	})
}

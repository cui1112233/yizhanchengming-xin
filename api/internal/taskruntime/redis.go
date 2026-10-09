package taskruntime

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrQueueEmpty = errors.New("runtime queue empty")
var ErrMalformedEnvelope = errors.New("runtime malformed queue envelope")

const processingReceiptTTL = 30 * time.Second
const queuePollInterval = 50 * time.Millisecond

type queueEnvelope struct {
	Receipt string  `json:"receipt"`
	Message Message `json:"message"`
}

type RedisQueue struct {
	client *redis.Client
	prefix string
}

func NewRedisQueue(addr, prefix string) (*RedisQueue, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, ErrQueueUnavailable
	}
	c := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("%w: %v", ErrQueueUnavailable, err)
	}
	return &RedisQueue{client: c, prefix: prefix}, nil
}
func (q *RedisQueue) Close() error          { return q.client.Close() }
func (q *RedisQueue) readyKey() string      { return q.prefix + ":queue:ready" }
func (q *RedisQueue) processingKey() string { return q.prefix + ":queue:processing" }
func (q *RedisQueue) delayedKey() string    { return q.prefix + ":queue:delayed" }

func (q *RedisQueue) processingDeadlineKey() string { return q.prefix + ":queue:processing:deadline" }
func (q *RedisQueue) processingEnvelopeKey() string { return q.prefix + ":queue:processing:envelope" }

// Moving a receipt and registering its deadline must be one operation: otherwise
// the upgrade pass could reclaim a delivery while its worker is registering it.
var claimDeliveryScript = redis.NewScript(`
local types={'list','list','hash','zset'}
for i,expected in ipairs(types) do
 local actual=redis.call('TYPE',KEYS[i]).ok
 if actual~='none' and actual~=expected then return redis.error_reply('WRONGTYPE queue key') end
end
local raw=redis.call('RPOP',KEYS[1])
if not raw then return {} end
local ok,env=pcall(cjson.decode,raw)
if not ok or type(env)~='table' or type(env.message)~='table' then return {raw,''} end
env.receipt=ARGV[1]
raw=cjson.encode(env)
redis.call('LPUSH',KEYS[2],raw)
redis.call('HSET',KEYS[3],ARGV[1],raw)
redis.call('ZADD',KEYS[4],ARGV[2],ARGV[1])
return {raw,ARGV[1]}`)

var settleDeliveryScript = redis.NewScript(`
local types={'list','hash','zset','list','zset'}
for i,expected in ipairs(types) do
 local actual=redis.call('TYPE',KEYS[i]).ok
 if actual~='none' and actual~=expected then return redis.error_reply('WRONGTYPE queue key') end
end
local raw=redis.call('HGET',KEYS[2],ARGV[1]) or ARGV[2]
local removed=redis.call('LREM',KEYS[1],1,raw)
redis.call('HDEL',KEYS[2],ARGV[1])
redis.call('ZREM',KEYS[3],ARGV[1])
if removed==1 and ARGV[3]=='nack' then
 if tonumber(ARGV[4])<=0 then redis.call('LPUSH',KEYS[4],raw)
 else redis.call('ZADD',KEYS[5],ARGV[4],raw) end
end
return removed`)

var reclaimProcessingScript = redis.NewScript(`
local types={'zset','hash','list','list'}
for i,expected in ipairs(types) do
 local actual=redis.call('TYPE',KEYS[i]).ok
 if actual~='none' and actual~=expected then return redis.error_reply('WRONGTYPE queue key') end
end
local limit=tonumber(ARGV[2])
local count=0
local expired=redis.call('ZRANGEBYSCORE',KEYS[1],'-inf',ARGV[1],'LIMIT',0,limit)
for _,receipt in ipairs(expired) do
 local raw=redis.call('HGET',KEYS[2],receipt)
 redis.call('ZREM',KEYS[1],receipt)
 redis.call('HDEL',KEYS[2],receipt)
 if raw and redis.call('LREM',KEYS[3],1,raw)==1 then
  redis.call('LPUSH',KEYS[4],raw)
  count=count+1
 end
end
-- Rotate the bounded scan so live entries cannot hide upgrade-era receipts.
local scans=math.min(limit-#expired,redis.call('LLEN',KEYS[3]))
for i=1,scans do
 local raw=redis.call('LPOP',KEYS[3])
 local ok,env=pcall(cjson.decode,raw)
 local receipt=ok and type(env)=='table' and env.receipt or nil
 local tracked=type(receipt)=='string' and redis.call('HGET',KEYS[2],receipt)==raw and redis.call('ZSCORE',KEYS[1],receipt)
 if tracked then redis.call('RPUSH',KEYS[3],raw)
 else
  if type(receipt)=='string' and redis.call('HGET',KEYS[2],receipt)==raw then
   redis.call('HDEL',KEYS[2],receipt);redis.call('ZREM',KEYS[1],receipt)
  end
  redis.call('LPUSH',KEYS[4],raw)
  count=count+1
 end
end
return count`)

func (q *RedisQueue) ReclaimExpired(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	count, err := reclaimProcessingScript.Run(ctx, q.client, []string{q.processingDeadlineKey(), q.processingEnvelopeKey(), q.processingKey(), q.readyKey()}, now.UnixMilli(), limit).Int()
	if err != nil {
		return 0, queueError(ctx, err)
	}
	return count, nil
}

func queueError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrQueueUnavailable, err)
}

func randomReceipt() (string,error) { b:=make([]byte,16); if _,err:=rand.Read(b);err!=nil{return "",err}; return hex.EncodeToString(b),nil }
func encodeDelivery(d Delivery)(string,error){ return encodeEnvelope(queueEnvelope{Receipt:d.Receipt,Message:d.Message}) }
func encodeEnvelope(e queueEnvelope)(string,error){ b,err:=json.Marshal(e); return string(b),err }

func (q *RedisQueue) Enqueue(ctx context.Context, msg Message) error {
	receipt,err:=randomReceipt(); if err!=nil{return err}
	raw,err:=encodeEnvelope(queueEnvelope{Receipt:receipt,Message:msg}); if err!=nil{return err}
	if err:=q.client.LPush(ctx,q.readyKey(),raw).Err();err!=nil{return fmt.Errorf("%w: %v",ErrQueueUnavailable,err)}
	return nil
}

var promoteDueScript = redis.NewScript(`
local types={'zset','list'}
for i,expected in ipairs(types) do
 local actual=redis.call('TYPE',KEYS[i]).ok
 if actual~='none' and actual~=expected then return redis.error_reply('WRONGTYPE queue key') end
end
local items=redis.call('ZRANGEBYSCORE',KEYS[1],'-inf',ARGV[1],'LIMIT',0,ARGV[2])
for _,v in ipairs(items) do
  if redis.call('ZREM',KEYS[1],v)==1 then redis.call('LPUSH',KEYS[2],v) end
end
return #items`)

func (q *RedisQueue) promoteDue(ctx context.Context) error {
	_,err:=promoteDueScript.Run(ctx,q.client,[]string{q.delayedKey(),q.readyKey()},time.Now().UnixMilli(),100).Result()
	if err!=nil{return fmt.Errorf("%w: %v",ErrQueueUnavailable,err)}
	return nil
}

func (q *RedisQueue) Claim(ctx context.Context, consumer string, wait time.Duration) (Delivery, error) {
	_ = consumer
	deadline := time.Now().Add(wait)
	for {
		if err := ctx.Err(); err != nil {
			return Delivery{}, err
		}
		if err := q.promoteDue(ctx); err != nil {
			return Delivery{}, queueError(ctx, err)
		}
		receipt, err := randomReceipt()
		if err != nil {
			return Delivery{}, err
		}
		values, err := claimDeliveryScript.Run(ctx, q.client, []string{q.readyKey(), q.processingKey(), q.processingEnvelopeKey(), q.processingDeadlineKey()}, receipt, time.Now().Add(processingReceiptTTL).UnixMilli()).StringSlice()
		if err != nil {
			return Delivery{}, queueError(ctx, err)
		}
		if len(values) != 0 {
			if values[1] == "" {
				return Delivery{}, ErrMalformedEnvelope
			}
			var env queueEnvelope
			if err := json.Unmarshal([]byte(values[0]), &env); err != nil {
				if ackErr := q.Ack(ctx, Delivery{Receipt: values[1]}); ackErr != nil {
					return Delivery{}, ackErr
				}
				return Delivery{}, ErrMalformedEnvelope
			}
			return Delivery{Message: env.Message, Receipt: env.Receipt}, nil
		}
		delay := queuePollInterval
		if wait > 0 {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return Delivery{}, ErrQueueEmpty
			}
			if remaining < delay {
				delay = remaining
			}
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Delivery{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func (q *RedisQueue) Ack(ctx context.Context, d Delivery) error {
	return q.settle(ctx, d, "ack", 0)
}

func (q *RedisQueue) Nack(ctx context.Context, d Delivery, delay time.Duration) error {
	return q.settle(ctx, d, "nack", delay)
}

func (q *RedisQueue) settle(ctx context.Context, d Delivery, action string, delay time.Duration) error {
	raw, err := encodeDelivery(d)
	if err != nil {
		return err
	}
	var due int64
	if delay > 0 {
		due = time.Now().Add(delay).UnixMilli()
	}
	_, err = settleDeliveryScript.Run(ctx, q.client, []string{q.processingKey(), q.processingEnvelopeKey(), q.processingDeadlineKey(), q.readyKey(), q.delayedKey()}, d.Receipt, raw, action, due).Result()
	if err != nil {
		return queueError(ctx, err)
	}
	return nil
}

func (q *RedisQueue) Flush(ctx context.Context) error {
	var cursor uint64
	for {
		keys,next,err:=q.client.Scan(ctx,cursor,q.prefix+"*",100).Result();if err!=nil{return err}
		if len(keys)>0 {if err:=q.client.Del(ctx,keys...).Err();err!=nil{return err}}
		cursor=next;if cursor==0{return nil}
	}
}

type RedisLeaseStore struct { client *redis.Client; prefix string }
func NewRedisLeaseStore(addr,prefix string)(*RedisLeaseStore,error){
	if strings.TrimSpace(addr)==""{return nil,ErrLeaseUnavailable}
	c:=redis.NewClient(&redis.Options{Addr:addr});ctx,cancel:=context.WithTimeout(context.Background(),2*time.Second);defer cancel();if err:=c.Ping(ctx).Err();err!=nil{_=c.Close();return nil,fmt.Errorf("%w: %v",ErrLeaseUnavailable,err)}
	return &RedisLeaseStore{client:c,prefix:prefix},nil
}
func (s *RedisLeaseStore) Close() error{return s.client.Close()}
func taskHash(task string)string{sum:=sha256.Sum256([]byte(task));return hex.EncodeToString(sum[:])}
func (s *RedisLeaseStore) leaseKey(hash string)string{return s.prefix+":lease:"+hash}
func (s *RedisLeaseStore) expiryKey()string{return s.prefix+":leases:expiry"}
func (s *RedisLeaseStore) metaKey()string{return s.prefix+":leases:meta"}
func leaseValue(owner string,token uint64)string{return owner+"\x1f"+strconv.FormatUint(token,10)}
func leaseMeta(task string,token uint64)string{return task+"\x1f"+strconv.FormatUint(token,10)}

var leaseClaimScript=redis.NewScript(`
if redis.call('SET',KEYS[1],ARGV[1],'NX','PX',ARGV[2]) then
 redis.call('HSET',KEYS[2],ARGV[3],ARGV[4])
 redis.call('ZADD',KEYS[3],ARGV[5],ARGV[3])
 return 1
end
return 0`)
var leaseRenewScript=redis.NewScript(`
if redis.call('GET',KEYS[1])==ARGV[1] then
 redis.call('PEXPIRE',KEYS[1],ARGV[2])
 redis.call('ZADD',KEYS[2],ARGV[3],ARGV[4])
 return 1
end
return 0`)
var leaseReleaseScript=redis.NewScript(`
if redis.call('GET',KEYS[1])==ARGV[1] then
 redis.call('DEL',KEYS[1])
 redis.call('ZREM',KEYS[2],ARGV[2])
 redis.call('HDEL',KEYS[3],ARGV[2])
 return 1
end
return 0`)
var expiredLeaseScript=redis.NewScript(`
local members=redis.call('ZRANGEBYSCORE',KEYS[1],'-inf',ARGV[1],'LIMIT',0,ARGV[2])
local out={}
for _,m in ipairs(members) do
 local lk=ARGV[3]..m
 if redis.call('EXISTS',lk)==0 then
  local meta=redis.call('HGET',KEYS[2],m)
  redis.call('ZREM',KEYS[1],m)
  redis.call('HDEL',KEYS[2],m)
  if meta then table.insert(out,meta) end
 end
end
return out`)

func (s *RedisLeaseStore) Claim(ctx context.Context,task,owner string,token uint64,ttl time.Duration)(Lease,bool,error){
	if ttl<=0{return Lease{},false,ErrLeaseUnavailable}
	h:=taskHash(task);exp:=time.Now().Add(ttl)
	v,err:=leaseClaimScript.Run(ctx,s.client,[]string{s.leaseKey(h),s.metaKey(),s.expiryKey()},leaseValue(owner,token),ttl.Milliseconds(),h,leaseMeta(task,token),exp.UnixMilli()).Int()
	if err!=nil{return Lease{},false,fmt.Errorf("%w: %v",ErrLeaseUnavailable,err)}
	return Lease{TaskKey:task,Owner:owner,FencingToken:token,ExpiresAt:exp},v==1,nil
}
func (s *RedisLeaseStore) Renew(ctx context.Context,l Lease,ttl time.Duration)(bool,error){
	h:=taskHash(l.TaskKey);exp:=time.Now().Add(ttl);v,err:=leaseRenewScript.Run(ctx,s.client,[]string{s.leaseKey(h),s.expiryKey()},leaseValue(l.Owner,l.FencingToken),ttl.Milliseconds(),exp.UnixMilli(),h).Int();if err!=nil{return false,fmt.Errorf("%w: %v",ErrLeaseUnavailable,err)};if v!=1{return false,ErrLeaseNotOwner};return true,nil
}
func (s *RedisLeaseStore) Release(ctx context.Context,l Lease)(bool,error){
	h:=taskHash(l.TaskKey);v,err:=leaseReleaseScript.Run(ctx,s.client,[]string{s.leaseKey(h),s.expiryKey(),s.metaKey()},leaseValue(l.Owner,l.FencingToken),h).Int();if err!=nil{return false,fmt.Errorf("%w: %v",ErrLeaseUnavailable,err)};if v!=1{return false,ErrLeaseNotOwner};return true,nil
}
func (s *RedisLeaseStore) RequeueExpired(ctx context.Context,now time.Time,limit int)([]ExpiredLease,error){
	if limit<=0{return nil,nil};raw,err:=expiredLeaseScript.Run(ctx,s.client,[]string{s.expiryKey(),s.metaKey()},now.UnixMilli(),limit,s.prefix+":lease:").Result();if err!=nil{return nil,fmt.Errorf("%w: %v",ErrLeaseUnavailable,err)}
	vals,ok:=raw.([]interface{});if !ok{return nil,nil};out:=make([]ExpiredLease,0,len(vals));for _,v:=range vals{str:=fmt.Sprint(v);idx:=strings.LastIndex(str,"\x1f");if idx<0{continue};tok,err:=strconv.ParseUint(str[idx+1:],10,64);if err!=nil{continue};out=append(out,ExpiredLease{TaskKey:str[:idx],FencingToken:tok})};return out,nil
}

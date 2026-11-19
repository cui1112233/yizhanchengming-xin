"""Read-only Nuitka constant stream decoder. No pyd import, marshal, pickle, eval or code execution."""
from pathlib import Path
import struct,json,hashlib,argparse,re
ROOT=Path(__file__).resolve().parent
TARGET=ROOT.parents[2]/'_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd'
args=argparse.ArgumentParser();args.add_argument('--module',default='novels.services.post_write_validator');args.add_argument('--slug',default='postwrite');args.add_argument('--redact',action='store_true');args.add_argument('--module-offset',type=int);options=args.parse_args()
b=TARGET.read_bytes();name=options.module.encode('ascii');off=options.module_offset if options.module_offset is not None else b.index(name+b'\0',3879424)
header=off+len(name)+1;size=struct.unpack_from('<I',b,header)[0];start=header+4;end=start+size
p=start+2;count=struct.unpack_from('<H',b,start)[0]
def varint():
 global p
 out=0;shift=0
 while True:
  v=b[p];p+=1;out|=(v&127)<<shift
  if v<128:return out
  shift+=7
def cstr():
 global p
 q=b.index(0,p,end);s=b[p:q].decode('utf8');p=q+1;return s
def many(n):
 out=[]
 for _ in range(n):out.append(one(out[-1] if out else None))
 return out
def one(previous=None):
 global p
 at=p;tag=chr(b[p]);p+=1
 if tag in 'au':value=cstr()
 elif tag in 'wv':
  n=1 if tag=='w' else varint();value=b[p:p+n].decode('utf8');p+=n
 elif tag=='s':value=''
 elif tag in 'nFt':value={'n':None,'F':False,'t':True}[tag]
 elif tag=='p':value={'previous':previous}
 elif tag in 'lq':value=varint()*(1 if tag=='l' else -1)
 elif tag=='f':value=struct.unpack_from('<d',b,p)[0];p+=8
 elif tag=='Z':value={'type':'special-float','code':b[p]};p+=1
 elif tag in 'TLSP':value={'type':{'T':'tuple','L':'list','S':'set','P':'frozenset'}[tag],'items':many(varint())}
 elif tag=='D':
  n=varint();value={'type':'dict','keys':many(n),'values':many(n)}
 elif tag in ':;':value={'type':tag,'items':many(3)}
 elif tag in 'MOEQ':
  if tag in 'OE':value={'builtinTag':tag,'name':cstr()}
  else:value={'builtinTag':tag,'index':b[p]};p+=1
 elif tag=='C':
  flags=varint();fn=one();line=varint()+1;args=one();argc=varint();qual=one() if flags&1 else fn;free=one() if flags&2 else None;kw=varint()+1 if flags&4 else 0;pos=varint()+1 if flags&8 else 0
  value={'type':'code-metadata-only','flags':flags,'name':fn,'line':line,'args':args,'argCount':argc,'qualname':qual,'freevars':free,'kwOnly':kw,'posOnly':pos}
 else:raise ValueError(f'unsupported tag {tag!r} at {at}')
 return {'tag':tag,'fileRange':[at,p],'value':value}
entries=[];error=None
try:
 for i in range(count):
  n=one(entries[-1] if entries else None);n['index']=i;entries.append(n)
except Exception as e:error=str(e)
out=ROOT/('constants/'+options.slug+'.json')
result={'moduleName':options.module,'targetSha256':hashlib.sha256(b).hexdigest(),'moduleNameOffset':off,'payloadRange':[start,end],'payloadSha256':hashlib.sha256(b[start:end]).hexdigest(),'declaredConstants':count,'parsedConstants':len(entries),'stoppedAt':p,'trailingHex':b[p:end].hex(),'complete':len(entries)==count and b[p:end]==b'.','error':error,'formatReference':'https://github.com/Nuitka/Nuitka/blob/2.7.12/nuitka/build/static_src/HelpersConstantsBlob.c','pythonVersionEvidence':'python312.dll literal in target; code metadata optional fields decoded for Python 3.12','entries':entries}
def redact(value):
 if isinstance(value,list):return [redact(x)for x in value]
 if isinstance(value,dict):
  r={k:redact(v)for k,v in value.items()}
  if value.get('tag')in['a','u','w','v'] and isinstance(value.get('value'),str):
   s=value['value']
   if not(len(s)<=140 and re.fullmatch(r'[A-Za-z_0-9.<>:\\-]*',s)):
    r['value']={'redacted':True,'chars':len(s),'sha256':hashlib.sha256(s.encode('utf8')).hexdigest()}
  return r
 return value
if options.redact:
 result=redact(result);result['redaction']='Only bounded ASCII identifiers are retained as strings; all prose, prompts and other string literals are omitted with length/hash metadata.'
 if len(b[p:end])>1:result['trailingHex']={'redacted':True,'bytes':len(b[p:end]),'sha256':hashlib.sha256(b[p:end]).hexdigest()}
out.write_text(json.dumps(result,ensure_ascii=False,indent=2),encoding='utf8')
print(json.dumps({k:result[k] for k in ['moduleNameOffset','payloadRange','declaredConstants','parsedConstants','stoppedAt','complete','error']}))
for e in []:
 if isinstance(e['value'],str) and (e['value'].startswith('_') or e['value']=='validate_chapter' or e['value']=='PostWriteValidationResult'):print(e['index'],e['fileRange'],repr(e['value']))

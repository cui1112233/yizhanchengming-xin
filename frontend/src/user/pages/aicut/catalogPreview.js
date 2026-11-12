import { aicutTextResourcePreset } from '../aicutText.js';

export function catalogTextSettings(id, parameters = {}) {
 return { ...(aicutTextResourcePreset(id) || {}), fontSize: Math.max(12,Math.min(240,Number(parameters.fontSize ?? 48))), textOpacity: Math.max(0,Math.min(1,Number(parameters.opacity ?? 1))) };
}
export function gradientGlyphColor(index, count) {
 const stops=[[255,214,107],[183,139,255],[99,220,255]];
 const position=count>1?Math.max(0,Math.min(1,index/(count-1)))*2:0;
 const left=Math.min(1,Math.floor(position)),ratio=position-left;
 return '#'+stops[left].map((value,c)=>Math.round(value+(stops[left+1][c]-value)*ratio).toString(16).padStart(2,'0')).join('');
}

// A resource can contain multiple actual passes; no generic fallback is shown as a named effect.
export function catalogPreviewPasses(id, parameters = {}) {
 const i=Math.max(0,Math.min(1,Number(parameters.intensity ?? .5))),speed=Math.max(.1,Number(parameters.speed ?? 1));
 if (!i || !id || id==='filter-original') return [];
 const pass=(mode,values=[])=>({mode,values,intensity:i,speed});
 const eq=(b=0,c=1,s=1,g=1)=>pass(13,[b,c,s,g]);
 const balance=(r,g,b)=>pass(14,[r,g,b]);
 switch(id) {
 case 'effect-blur': return [pass(1,[Math.min(20,Number(parameters.radius)||2+8*i)])];
 case 'effect-vignette':return [pass(2,[.2+.8*i])];
 case 'effect-glow':return [pass(1,[.5+2.5*i])];
 case 'effect-grain':return [pass(4,[(2+Math.trunc(18*i))/255])];
 case 'effect-shake':return [pass(5,[2+Math.trunc(12*i)])];
 case 'effect-breathe-zoom':return [pass(6)];
 case 'effect-flash':return [pass(7)];
 case 'effect-glitch':return [pass(8,[1+Math.trunc(8*i)])];
 case 'effect-chromatic':return [pass(9,[1+Math.trunc(6*i)])];
 case 'effect-mosaic':return [pass(10,[4+Math.trunc(28*i)])];
 case 'effect-fisheye':return [pass(11,[-.2*i,.05*i])];
 case 'effect-sharpen':return [pass(12,[.5+1.5*i])];
 case 'filter-bright':return [eq(.12*i,1,1+.08*i)];
 case 'filter-clear':return [pass(12,[.5+i])];
 case 'filter-natural':return [eq(0,1+.04*i,1+.04*i)];
 case 'filter-warm-sun':return [balance(.12*i,.04*i,-.08*i)];
 case 'filter-cool-blue':return [balance(-.08*i,0,.14*i)];
 case 'filter-cinema':return [eq(0,1+.18*i,1-.12*i),pass(2,[.35])];
 case 'filter-retro':return [balance(.10*i,.05*i,-.10*i),pass(4,[4/255])];
 case 'filter-monochrome':return [eq(0,1,0)];
 case 'filter-high-contrast':return [eq(0,1+.45*i)];
 case 'filter-low-saturation':return [eq(0,1,1-.7*i)];
 case 'filter-teal-orange':return [balance(.12*i,-.03*i,.10*i)];
 case 'filter-night':return [eq(.04*i,1,1-.15*i,1+.22*i)];
 case 'filter-portrait':return [eq(.04*i,1,1+.06*i),pass(1,[.15+.25*i])];
 case 'filter-vivid':return [eq(0,1+.12*i,1+.45*i)];
 case 'filter-soft':return [eq(0,1-.16*i),pass(1,[.25+.45*i])];
 case 'filter-golden-hour':return [balance(.16*i,.08*i,-.10*i),eq(.05*i,1+.08*i,1+.3*i)];
 case 'filter-noir':return [eq(-.03*i,1+.5*i,0,1-.1*i)];
 case 'filter-vintage':return [balance(.11*i,.05*i,-.09*i),eq(0,1-.1*i,1-.28*i),pass(4,[(2+Math.trunc(5*i))/255])];
 case 'filter-cyberpunk':return [balance(.16*i,-.1*i,.22*i),eq(0,1+.32*i,1+.75*i)];
 case 'filter-vaporwave':return [balance(.2*i,-.08*i,.18*i),eq(.05*i,1,1+.6*i,1+.1*i)];
 case 'filter-mint-haze':return [balance(-.06*i,.1*i,.08*i),eq(.07*i,1-.14*i,1-.12*i)];
 case 'filter-bleach-bypass':return [eq(-.02*i,1+.48*i,1-.55*i),pass(12,[.35+.55*i])];
 case 'filter-winter-cool':return [balance(-.1*i,.02*i,.18*i),eq(0,1+.1*i,1-.25*i)];
 case 'filter-rosy':return [balance(.14*i,-.03*i,.05*i),eq(.05*i,1,1+.35*i)];
 case 'filter-film':return [eq(-.01*i,1+.2*i,1-.12*i),pass(4,[(2+Math.trunc(7*i))/255]),pass(2,[.12+.18*i])];
 case 'filter-super-8':return [balance(.1*i,.03*i,-.08*i),eq(0,1+.28*i,1+.18*i),pass(4,[(4+Math.trunc(10*i))/255]),pass(2,[.18+.22*i])];
 case 'filter-instant-print':return [balance(.07*i,.04*i,-.04*i),eq(.08*i,1-.16*i,1-.18*i),pass(2,[.08+.1*i])];
 default:return [];
 }
}

const fragmentSource=`precision highp float;
varying vec2 uv;uniform sampler2D image;uniform vec2 size;uniform int mode;uniform vec4 p;uniform float intensity;uniform float speed;uniform float time;
float hash(vec2 q){return fract(sin(dot(q,vec2(12.9898,78.233)))*43758.5453);}
vec4 sampleAt(vec2 q){return texture2D(image,clamp(q,vec2(0.0),vec2(1.0)));}
void main(){vec2 q=uv;vec2 stepSize=1.0/size;
 if(mode==5){q+=vec2(p.x*sin(time*speed*10.0),p.x*cos(time*speed*13.0))/size;}
 if(mode==6){float z=min(1.0+.02*intensity*(1.0+sin(time*30.0*speed*.12)),1.1);q=(q-.5)/z+.5;}
 if(mode==10){q=(floor(q*size/p.x)+.5)*p.x/size;}
 if(mode==11){vec2 d=(q-.5)*2.0;float r2=dot(d,d);q=.5+d*(1.0+p.x*r2+p.y*r2*r2)*.5;}
 vec4 c=sampleAt(q);
 if(mode==1 || mode==12){vec2 d=stepSize*max(1.0,mode==1?p.x:1.0);vec4 avg=(sampleAt(q+vec2(d.x,0.0))+sampleAt(q-vec2(d.x,0.0))+sampleAt(q+vec2(0.0,d.y))+sampleAt(q-vec2(0.0,d.y)))*.25;c=mode==1?mix(c,avg,.8):vec4(c.rgb+(c.rgb-avg.rgb)*p.x,c.a);}
 if(mode==2){vec2 d=(q-.5)*2.0;c.rgb*=pow(max(.0,cos(min(1.55,length(d)*p.x))),4.0);}
 if(mode==4){c.rgb+=(hash(floor(q*size)+floor(time*30.0))-.5)*p.x*2.0;}
 if(mode==7){c.rgb+=.25*intensity*(1.0+sin(time*speed*12.0));}
 if(mode==8 || mode==9){c.r=sampleAt(q+vec2(p.x/size.x,0.0)).r;c.b=sampleAt(q+(mode==8?vec2(0.0,-p.x/size.y):vec2(-p.x/size.x,0.0))).b;}
 if(mode==13){vec3 rgb=(c.rgb-.5)*p.y+.5+p.x;float y=dot(rgb,vec3(.299,.587,.114));rgb=mix(vec3(y),rgb,p.z);c.rgb=pow(max(vec3(0.0),rgb),vec3(1.0/max(.1,p.w)));}
 if(mode==14){vec3 weight=1.0-smoothstep(vec3(0.0),vec3(1.0),c.rgb);c.rgb+=p.rgb*weight;}
 gl_FragColor=vec4(clamp(c.rgb,0.0,1.0),c.a);}`;

export function startCatalogPreviewRenderer(canvas, source, initialLayers = [], initialTime = 0) {
 const gl=canvas?.getContext?.('webgl',{alpha:true,antialias:false,preserveDrawingBuffer:true});
 if(!gl || !source)return null;
 const shader=(kind,text)=>{const s=gl.createShader(kind);gl.shaderSource(s,text);gl.compileShader(s);if(!gl.getShaderParameter(s,gl.COMPILE_STATUS))throw new Error(gl.getShaderInfoLog(s));return s;};
 let program,vertex,fragment;
 try {vertex=shader(gl.VERTEX_SHADER,'attribute vec2 position;varying vec2 uv;void main(){uv=(position+1.0)*.5;gl_Position=vec4(position,0.,1.);}');fragment=shader(gl.FRAGMENT_SHADER,fragmentSource);program=gl.createProgram();gl.attachShader(program,vertex);gl.attachShader(program,fragment);gl.linkProgram(program);if(!gl.getProgramParameter(program,gl.LINK_STATUS))throw new Error('效果预览初始化失败');}catch{return null;}
 gl.useProgram(program);const buffer=gl.createBuffer();gl.bindBuffer(gl.ARRAY_BUFFER,buffer);gl.bufferData(gl.ARRAY_BUFFER,new Float32Array([-1,-1,1,-1,-1,1,-1,1,1,-1,1,1]),gl.STATIC_DRAW);const pos=gl.getAttribLocation(program,'position');gl.enableVertexAttribArray(pos);gl.vertexAttribPointer(pos,2,gl.FLOAT,false,0,0);
 const textures=[0,1,2].map(()=>{const t=gl.createTexture();gl.bindTexture(gl.TEXTURE_2D,t);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_MIN_FILTER,gl.LINEAR);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_MAG_FILTER,gl.LINEAR);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_WRAP_S,gl.CLAMP_TO_EDGE);gl.texParameteri(gl.TEXTURE_2D,gl.TEXTURE_WRAP_T,gl.CLAMP_TO_EDGE);return t;});
 const fbos=[gl.createFramebuffer(),gl.createFramebuffer()];let layers=initialLayers,seconds=initialTime/1000,stopped=false,frame=0,width=0,height=0;
 const loc=name=>gl.getUniformLocation(program,name);
 const draw=()=>{if(stopped)return;const w=source.videoWidth||source.naturalWidth||source.width,h=source.videoHeight||source.naturalHeight||source.height;if(w&&h){
  const scale=Math.min(1,960/w);const tw=Math.max(1,Math.round(w*scale)),th=Math.max(1,Math.round(h*scale));
  if(width!==tw||height!==th){width=tw;height=th;canvas.width=width;canvas.height=height;for(let n=0;n<2;n++){gl.bindTexture(gl.TEXTURE_2D,textures[n+1]);gl.texImage2D(gl.TEXTURE_2D,0,gl.RGBA,width,height,0,gl.RGBA,gl.UNSIGNED_BYTE,null);gl.bindFramebuffer(gl.FRAMEBUFFER,fbos[n]);gl.framebufferTexture2D(gl.FRAMEBUFFER,gl.COLOR_ATTACHMENT0,gl.TEXTURE_2D,textures[n+1],0);}}
  gl.bindTexture(gl.TEXTURE_2D,textures[0]);gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL,true);try{gl.texImage2D(gl.TEXTURE_2D,0,gl.RGBA,gl.RGBA,gl.UNSIGNED_BYTE,source);}catch{frame=requestAnimationFrame(draw);return;}
  const passes=layers.flatMap(layer=>catalogPreviewPasses(layer.resourceId,layer.parameters));if(!passes.length)passes.push({mode:0,values:[],intensity:0,speed:1});
  let input=textures[0];gl.viewport(0,0,width,height);gl.uniform2f(loc('size'),width,height);gl.uniform1f(loc('time'),seconds);
  passes.forEach((pass,n)=>{const last=n===passes.length-1;gl.bindFramebuffer(gl.FRAMEBUFFER,last?null:fbos[n%2]);gl.bindTexture(gl.TEXTURE_2D,input);gl.uniform1i(loc('mode'),pass.mode);gl.uniform4fv(loc('p'),new Float32Array([...pass.values,0,0,0,0].slice(0,4)));gl.uniform1f(loc('intensity'),pass.intensity);gl.uniform1f(loc('speed'),pass.speed);gl.drawArrays(gl.TRIANGLES,0,6);input=textures[n%2+1];});
 }frame=requestAnimationFrame(draw);};draw();
 return {update(next,time){layers=next;seconds=time/1000;},stop(){stopped=true;cancelAnimationFrame(frame);textures.forEach(t=>gl.deleteTexture(t));fbos.forEach(f=>gl.deleteFramebuffer(f));gl.deleteBuffer(buffer);gl.deleteProgram(program);gl.deleteShader(vertex);gl.deleteShader(fragment);}};
}

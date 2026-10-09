var oe=Object.defineProperty;var re=(t,e,o)=>e in t?oe(t,e,{enumerable:!0,configurable:!0,writable:!0,value:o}):t[e]=o;var l=(t,e,o)=>re(t,typeof e!="symbol"?e+"":e,o);import{r as U,j as G}from"./createLucideIcon-Csms46Jc.js";const se=`#version 300 es
precision mediump float;

layout(location = 0) in vec4 a_position;

uniform vec2 u_resolution;
uniform float u_pixelRatio;
uniform float u_imageAspectRatio;
uniform float u_originX;
uniform float u_originY;
uniform float u_worldWidth;
uniform float u_worldHeight;
uniform float u_fit;
uniform float u_scale;
uniform float u_rotation;
uniform float u_offsetX;
uniform float u_offsetY;

out vec2 v_objectUV;
out vec2 v_objectBoxSize;
out vec2 v_responsiveUV;
out vec2 v_responsiveBoxGivenSize;
out vec2 v_patternUV;
out vec2 v_patternBoxSize;
out vec2 v_imageUV;

vec3 getBoxSize(float boxRatio, vec2 givenBoxSize) {
  vec2 box = vec2(0.);
  // fit = none
  box.x = boxRatio * min(givenBoxSize.x / boxRatio, givenBoxSize.y);
  float noFitBoxWidth = box.x;
  if (u_fit == 1.) { // fit = contain
    box.x = boxRatio * min(u_resolution.x / boxRatio, u_resolution.y);
  } else if (u_fit == 2.) { // fit = cover
    box.x = boxRatio * max(u_resolution.x / boxRatio, u_resolution.y);
  }
  box.y = box.x / boxRatio;
  return vec3(box, noFitBoxWidth);
}

void main() {
  gl_Position = a_position;

  vec2 uv = gl_Position.xy * .5;
  vec2 boxOrigin = vec2(.5 - u_originX, u_originY - .5);
  vec2 givenBoxSize = vec2(u_worldWidth, u_worldHeight);
  givenBoxSize = max(givenBoxSize, vec2(1.)) * u_pixelRatio;
  float r = u_rotation * 3.14159265358979323846 / 180.;
  mat2 graphicRotation = mat2(cos(r), sin(r), -sin(r), cos(r));
  vec2 graphicOffset = vec2(-u_offsetX, u_offsetY);


  // ===================================================

  float fixedRatio = 1.;
  vec2 fixedRatioBoxGivenSize = vec2(
  (u_worldWidth == 0.) ? u_resolution.x : givenBoxSize.x,
  (u_worldHeight == 0.) ? u_resolution.y : givenBoxSize.y
  );

  v_objectBoxSize = getBoxSize(fixedRatio, fixedRatioBoxGivenSize).xy;
  vec2 objectWorldScale = u_resolution.xy / v_objectBoxSize;

  v_objectUV = uv;
  v_objectUV *= objectWorldScale;
  v_objectUV += boxOrigin * (objectWorldScale - 1.);
  v_objectUV += graphicOffset;
  v_objectUV /= u_scale;
  v_objectUV = graphicRotation * v_objectUV;

  // ===================================================

  v_responsiveBoxGivenSize = vec2(
  (u_worldWidth == 0.) ? u_resolution.x : givenBoxSize.x,
  (u_worldHeight == 0.) ? u_resolution.y : givenBoxSize.y
  );
  float responsiveRatio = v_responsiveBoxGivenSize.x / v_responsiveBoxGivenSize.y;
  vec2 responsiveBoxSize = getBoxSize(responsiveRatio, v_responsiveBoxGivenSize).xy;
  vec2 responsiveBoxScale = u_resolution.xy / responsiveBoxSize;

  #ifdef ADD_HELPERS
  v_responsiveHelperBox = uv;
  v_responsiveHelperBox *= responsiveBoxScale;
  v_responsiveHelperBox += boxOrigin * (responsiveBoxScale - 1.);
  #endif

  v_responsiveUV = uv;
  v_responsiveUV *= responsiveBoxScale;
  v_responsiveUV += boxOrigin * (responsiveBoxScale - 1.);
  v_responsiveUV += graphicOffset;
  v_responsiveUV /= u_scale;
  v_responsiveUV.x *= responsiveRatio;
  v_responsiveUV = graphicRotation * v_responsiveUV;
  v_responsiveUV.x /= responsiveRatio;

  // ===================================================

  float patternBoxRatio = givenBoxSize.x / givenBoxSize.y;
  vec2 patternBoxGivenSize = vec2(
  (u_worldWidth == 0.) ? u_resolution.x : givenBoxSize.x,
  (u_worldHeight == 0.) ? u_resolution.y : givenBoxSize.y
  );
  patternBoxRatio = patternBoxGivenSize.x / patternBoxGivenSize.y;

  vec3 boxSizeData = getBoxSize(patternBoxRatio, patternBoxGivenSize);
  v_patternBoxSize = boxSizeData.xy;
  float patternBoxNoFitBoxWidth = boxSizeData.z;
  vec2 patternBoxScale = u_resolution.xy / v_patternBoxSize;

  v_patternUV = uv;
  v_patternUV += graphicOffset / patternBoxScale;
  v_patternUV += boxOrigin;
  v_patternUV -= boxOrigin / patternBoxScale;
  v_patternUV *= u_resolution.xy;
  v_patternUV /= u_pixelRatio;
  if (u_fit > 0.) {
    v_patternUV *= (patternBoxNoFitBoxWidth / v_patternBoxSize.x);
  }
  v_patternUV /= u_scale;
  v_patternUV = graphicRotation * v_patternUV;
  v_patternUV += boxOrigin / patternBoxScale;
  v_patternUV -= boxOrigin;
  // x100 is a default multiplier between vertex and fragmant shaders
  // we use it to avoid UV presision issues
  v_patternUV *= .01;

  // ===================================================

  vec2 imageBoxSize;
  if (u_fit == 1.) { // contain
    imageBoxSize.x = min(u_resolution.x / u_imageAspectRatio, u_resolution.y) * u_imageAspectRatio;
  } else if (u_fit == 2.) { // cover
    imageBoxSize.x = max(u_resolution.x / u_imageAspectRatio, u_resolution.y) * u_imageAspectRatio;
  } else {
    imageBoxSize.x = min(10.0, 10.0 / u_imageAspectRatio * u_imageAspectRatio);
  }
  imageBoxSize.y = imageBoxSize.x / u_imageAspectRatio;
  vec2 imageBoxScale = u_resolution.xy / imageBoxSize;

  v_imageUV = uv;
  v_imageUV *= imageBoxScale;
  v_imageUV += boxOrigin * (imageBoxScale - 1.);
  v_imageUV += graphicOffset;
  v_imageUV /= u_scale;
  v_imageUV.x *= u_imageAspectRatio;
  v_imageUV = graphicRotation * v_imageUV;
  v_imageUV.x /= u_imageAspectRatio;

  v_imageUV += .5;
  v_imageUV.y = 1. - v_imageUV.y;
}`,N=1920*1080*4;let ae=class{constructor(e,o,i,n,s=0,a=0,r=2,f=N,c=[]){l(this,"parentElement");l(this,"canvasElement");l(this,"gl");l(this,"program",null);l(this,"uniformLocations",{});l(this,"fragmentShader");l(this,"rafId",null);l(this,"lastRenderTime",0);l(this,"currentFrame",0);l(this,"speed",0);l(this,"currentSpeed",0);l(this,"providedUniforms");l(this,"mipmaps",[]);l(this,"hasBeenDisposed",!1);l(this,"resolutionChanged",!0);l(this,"textures",new Map);l(this,"minPixelRatio");l(this,"maxPixelCount");l(this,"isSafari",ce());l(this,"uniformCache",{});l(this,"textureUnitMap",new Map);l(this,"ownerDocument");l(this,"initProgram",()=>{const e=ne(this.gl,se,this.fragmentShader);e&&(this.program=e)});l(this,"setupPositionAttribute",()=>{const e=this.gl.getAttribLocation(this.program,"a_position"),o=this.gl.createBuffer();this.gl.bindBuffer(this.gl.ARRAY_BUFFER,o);const i=[-1,-1,1,-1,-1,1,-1,1,1,-1,1,1];this.gl.bufferData(this.gl.ARRAY_BUFFER,new Float32Array(i),this.gl.STATIC_DRAW),this.gl.enableVertexAttribArray(e),this.gl.vertexAttribPointer(e,2,this.gl.FLOAT,!1,0,0)});l(this,"setupUniforms",()=>{const e={u_time:this.gl.getUniformLocation(this.program,"u_time"),u_pixelRatio:this.gl.getUniformLocation(this.program,"u_pixelRatio"),u_resolution:this.gl.getUniformLocation(this.program,"u_resolution")};Object.entries(this.providedUniforms).forEach(([o,i])=>{if(e[o]=this.gl.getUniformLocation(this.program,o),i instanceof HTMLImageElement){const n=`${o}AspectRatio`;e[n]=this.gl.getUniformLocation(this.program,n)}}),this.uniformLocations=e});l(this,"renderScale",1);l(this,"parentWidth",0);l(this,"parentHeight",0);l(this,"parentDevicePixelWidth",0);l(this,"parentDevicePixelHeight",0);l(this,"devicePixelsSupported",!1);l(this,"resizeObserver",null);l(this,"setupResizeObserver",()=>{this.resizeObserver=new ResizeObserver(([e])=>{var o;if(e!=null&&e.borderBoxSize[0]){const i=(o=e.devicePixelContentBoxSize)==null?void 0:o[0];i!==void 0&&(this.devicePixelsSupported=!0,this.parentDevicePixelWidth=i.inlineSize,this.parentDevicePixelHeight=i.blockSize),this.parentWidth=e.borderBoxSize[0].inlineSize,this.parentHeight=e.borderBoxSize[0].blockSize}this.handleResize()}),this.resizeObserver.observe(this.parentElement)});l(this,"handleVisualViewportChange",()=>{var e;(e=this.resizeObserver)==null||e.disconnect(),this.setupResizeObserver()});l(this,"handleResize",()=>{let e=0,o=0;const i=Math.max(1,window.devicePixelRatio),n=(visualViewport==null?void 0:visualViewport.scale)??1;if(this.devicePixelsSupported){const d=Math.max(1,this.minPixelRatio/i);e=this.parentDevicePixelWidth*d*n,o=this.parentDevicePixelHeight*d*n}else{let d=Math.max(i,this.minPixelRatio)*n;if(this.isSafari){const h=ue(this.ownerDocument);d*=Math.max(1,h)}e=Math.round(this.parentWidth)*d,o=Math.round(this.parentHeight)*d}const s=Math.sqrt(this.maxPixelCount)/Math.sqrt(e*o),a=Math.min(1,s),r=Math.round(e*a),f=Math.round(o*a),c=r/Math.round(this.parentWidth);(this.canvasElement.width!==r||this.canvasElement.height!==f||this.renderScale!==c)&&(this.renderScale=c,this.canvasElement.width=r,this.canvasElement.height=f,this.resolutionChanged=!0,this.gl.viewport(0,0,this.gl.canvas.width,this.gl.canvas.height),this.render(performance.now()))});l(this,"render",e=>{if(this.hasBeenDisposed)return;if(this.program===null){console.warn("Tried to render before program or gl was initialized");return}const o=e-this.lastRenderTime;this.lastRenderTime=e,this.currentSpeed!==0&&(this.currentFrame+=o*this.currentSpeed),this.gl.clear(this.gl.COLOR_BUFFER_BIT),this.gl.useProgram(this.program),this.gl.uniform1f(this.uniformLocations.u_time,this.currentFrame*.001),this.resolutionChanged&&(this.gl.uniform2f(this.uniformLocations.u_resolution,this.gl.canvas.width,this.gl.canvas.height),this.gl.uniform1f(this.uniformLocations.u_pixelRatio,this.renderScale),this.resolutionChanged=!1),this.gl.drawArrays(this.gl.TRIANGLES,0,6),this.currentSpeed!==0?this.requestRender():this.rafId=null});l(this,"requestRender",()=>{this.rafId!==null&&cancelAnimationFrame(this.rafId),this.rafId=requestAnimationFrame(this.render)});l(this,"setTextureUniform",(e,o)=>{if(!o.complete||o.naturalWidth===0)throw new Error(`Paper Shaders: image for uniform ${e} must be fully loaded`);const i=this.textures.get(e);i&&this.gl.deleteTexture(i),this.textureUnitMap.has(e)||this.textureUnitMap.set(e,this.textureUnitMap.size);const n=this.textureUnitMap.get(e);this.gl.activeTexture(this.gl.TEXTURE0+n);const s=this.gl.createTexture();this.gl.bindTexture(this.gl.TEXTURE_2D,s),this.gl.texParameteri(this.gl.TEXTURE_2D,this.gl.TEXTURE_WRAP_S,this.gl.CLAMP_TO_EDGE),this.gl.texParameteri(this.gl.TEXTURE_2D,this.gl.TEXTURE_WRAP_T,this.gl.CLAMP_TO_EDGE),this.gl.texParameteri(this.gl.TEXTURE_2D,this.gl.TEXTURE_MIN_FILTER,this.gl.LINEAR),this.gl.texParameteri(this.gl.TEXTURE_2D,this.gl.TEXTURE_MAG_FILTER,this.gl.LINEAR),this.gl.texImage2D(this.gl.TEXTURE_2D,0,this.gl.RGBA,this.gl.RGBA,this.gl.UNSIGNED_BYTE,o),this.mipmaps.includes(e)&&(this.gl.generateMipmap(this.gl.TEXTURE_2D),this.gl.texParameteri(this.gl.TEXTURE_2D,this.gl.TEXTURE_MIN_FILTER,this.gl.LINEAR_MIPMAP_LINEAR));const a=this.gl.getError();if(a!==this.gl.NO_ERROR||s===null){console.error("Paper Shaders: WebGL error when uploading texture:",a);return}this.textures.set(e,s);const r=this.uniformLocations[e];if(r){this.gl.uniform1i(r,n);const f=`${e}AspectRatio`,c=this.uniformLocations[f];if(c){const d=o.naturalWidth/o.naturalHeight;this.gl.uniform1f(c,d)}}});l(this,"areUniformValuesEqual",(e,o)=>e===o?!0:Array.isArray(e)&&Array.isArray(o)&&e.length===o.length?e.every((i,n)=>this.areUniformValuesEqual(i,o[n])):!1);l(this,"setUniformValues",e=>{this.gl.useProgram(this.program),Object.entries(e).forEach(([o,i])=>{let n=i;if(i instanceof HTMLImageElement&&(n=`${i.src.slice(0,200)}|${i.naturalWidth}x${i.naturalHeight}`),this.areUniformValuesEqual(this.uniformCache[o],n))return;this.uniformCache[o]=n;const s=this.uniformLocations[o];if(!s){console.warn(`Uniform location for ${o} not found`);return}if(i instanceof HTMLImageElement)this.setTextureUniform(o,i);else if(Array.isArray(i)){let a=null,r=null;if(i[0]!==void 0&&Array.isArray(i[0])){const f=i[0].length;if(i.every(c=>c.length===f))a=i.flat(),r=f;else{console.warn(`All child arrays must be the same length for ${o}`);return}}else a=i,r=a.length;switch(r){case 2:this.gl.uniform2fv(s,a);break;case 3:this.gl.uniform3fv(s,a);break;case 4:this.gl.uniform4fv(s,a);break;case 9:this.gl.uniformMatrix3fv(s,!1,a);break;case 16:this.gl.uniformMatrix4fv(s,!1,a);break;default:console.warn(`Unsupported uniform array length: ${r}`)}}else typeof i=="number"?this.gl.uniform1f(s,i):typeof i=="boolean"?this.gl.uniform1i(s,i?1:0):console.warn(`Unsupported uniform type for ${o}: ${typeof i}`)})});l(this,"getCurrentFrame",()=>this.currentFrame);l(this,"setFrame",e=>{this.currentFrame=e,this.lastRenderTime=performance.now(),this.render(performance.now())});l(this,"setSpeed",(e=1)=>{this.speed=e,this.setCurrentSpeed(this.ownerDocument.hidden?0:e)});l(this,"setCurrentSpeed",e=>{this.currentSpeed=e,this.rafId===null&&e!==0&&(this.lastRenderTime=performance.now(),this.rafId=requestAnimationFrame(this.render)),this.rafId!==null&&e===0&&(cancelAnimationFrame(this.rafId),this.rafId=null)});l(this,"setMaxPixelCount",(e=N)=>{this.maxPixelCount=e,this.handleResize()});l(this,"setMinPixelRatio",(e=2)=>{this.minPixelRatio=e,this.handleResize()});l(this,"setUniforms",e=>{this.setUniformValues(e),this.providedUniforms={...this.providedUniforms,...e},this.render(performance.now())});l(this,"handleDocumentVisibilityChange",()=>{this.setCurrentSpeed(this.ownerDocument.hidden?0:this.speed)});l(this,"dispose",()=>{this.hasBeenDisposed=!0,this.rafId!==null&&(cancelAnimationFrame(this.rafId),this.rafId=null),this.gl&&this.program&&(this.textures.forEach(e=>{this.gl.deleteTexture(e)}),this.textures.clear(),this.gl.deleteProgram(this.program),this.program=null,this.gl.bindBuffer(this.gl.ARRAY_BUFFER,null),this.gl.bindBuffer(this.gl.ELEMENT_ARRAY_BUFFER,null),this.gl.bindRenderbuffer(this.gl.RENDERBUFFER,null),this.gl.bindFramebuffer(this.gl.FRAMEBUFFER,null),this.gl.getError()),this.resizeObserver&&(this.resizeObserver.disconnect(),this.resizeObserver=null),visualViewport==null||visualViewport.removeEventListener("resize",this.handleVisualViewportChange),this.ownerDocument.removeEventListener("visibilitychange",this.handleDocumentVisibilityChange),this.uniformLocations={},this.canvasElement.remove(),delete this.parentElement.paperShaderMount});if((e==null?void 0:e.nodeType)===1)this.parentElement=e;else throw new Error("Paper Shaders: parent element must be an HTMLElement");if(this.ownerDocument=e.ownerDocument,!this.ownerDocument.querySelector("style[data-paper-shader]")){const R=this.ownerDocument.createElement("style");R.innerHTML=le,R.setAttribute("data-paper-shader",""),this.ownerDocument.head.prepend(R)}const d=this.ownerDocument.createElement("canvas");this.canvasElement=d,this.parentElement.prepend(d),this.fragmentShader=o,this.providedUniforms=i,this.mipmaps=c,this.currentFrame=a,this.minPixelRatio=r,this.maxPixelCount=f;const h=d.getContext("webgl2",n);if(!h)throw new Error("Paper Shaders: WebGL is not supported in this browser");this.gl=h,this.initProgram(),this.setupPositionAttribute(),this.setupUniforms(),this.setUniformValues(this.providedUniforms),this.setupResizeObserver(),visualViewport==null||visualViewport.addEventListener("resize",this.handleVisualViewportChange),this.setSpeed(s),this.parentElement.setAttribute("data-paper-shader",""),this.parentElement.paperShaderMount=this,this.ownerDocument.addEventListener("visibilitychange",this.handleDocumentVisibilityChange)}};function Y(t,e,o){const i=t.createShader(e);return i?(t.shaderSource(i,o),t.compileShader(i),t.getShaderParameter(i,t.COMPILE_STATUS)?i:(console.error("An error occurred compiling the shaders: "+t.getShaderInfoLog(i)),t.deleteShader(i),null)):null}function ne(t,e,o){const i=t.getShaderPrecisionFormat(t.FRAGMENT_SHADER,t.MEDIUM_FLOAT),n=i?i.precision:null;n&&n<23&&(e=e.replace(/precision\s+(lowp|mediump)\s+float;/g,"precision highp float;"),o=o.replace(/precision\s+(lowp|mediump)\s+float/g,"precision highp float").replace(/\b(uniform|varying|attribute)\s+(lowp|mediump)\s+(\w+)/g,"$1 highp $3"));const s=Y(t,t.VERTEX_SHADER,e),a=Y(t,t.FRAGMENT_SHADER,o);if(!s||!a)return null;const r=t.createProgram();return r?(t.attachShader(r,s),t.attachShader(r,a),t.linkProgram(r),t.getProgramParameter(r,t.LINK_STATUS)?(t.detachShader(r,s),t.detachShader(r,a),t.deleteShader(s),t.deleteShader(a),r):(console.error("Unable to initialize the shader program: "+t.getProgramInfoLog(r)),t.deleteProgram(r),t.deleteShader(s),t.deleteShader(a),null)):null}const le=`@layer paper-shaders {
  :where([data-paper-shader]) {
    isolation: isolate;
    position: relative;

    & canvas {
      contain: strict;
      display: block;
      position: absolute;
      inset: 0;
      z-index: -1;
      width: 100%;
      height: 100%;
      border-radius: inherit;
      corner-shape: inherit;
    }
  }
}`;function ce(){const t=navigator.userAgent.toLowerCase();return t.includes("safari")&&!t.includes("chrome")&&!t.includes("android")}function ue(t){const e=(visualViewport==null?void 0:visualViewport.scale)??1,o=(visualViewport==null?void 0:visualViewport.width)??window.innerWidth,i=window.innerWidth-t.documentElement.clientWidth,n=e*o+i,s=outerWidth/n,a=Math.round(100*s);return a%5===0?a/100:a===33?1/3:a===67?2/3:a===133?4/3:s}const he={fit:"contain",scale:1,rotation:0,offsetX:0,offsetY:0,originX:.5,originY:.5,worldWidth:0,worldHeight:0},de={none:0,contain:1,cover:2},fe=`
#define TWO_PI 6.28318530718
#define PI 3.14159265358979323846
`,pe=`
vec2 rotate(vec2 uv, float th) {
  return mat2(cos(th), sin(th), -sin(th), cos(th)) * uv;
}
`,me=`
  color += 1. / 256. * (fract(sin(dot(.014 * gl_FragCoord.xy, vec2(12.9898, 78.233))) * 43758.5453123) - .5);
`,ge=`
vec3 permute(vec3 x) { return mod(((x * 34.0) + 1.0) * x, 289.0); }
float snoise(vec2 v) {
  const vec4 C = vec4(0.211324865405187, 0.366025403784439,
    -0.577350269189626, 0.024390243902439);
  vec2 i = floor(v + dot(v, C.yy));
  vec2 x0 = v - i + dot(i, C.xx);
  vec2 i1;
  i1 = (x0.x > x0.y) ? vec2(1.0, 0.0) : vec2(0.0, 1.0);
  vec4 x12 = x0.xyxy + C.xxzz;
  x12.xy -= i1;
  i = mod(i, 289.0);
  vec3 p = permute(permute(i.y + vec3(0.0, i1.y, 1.0))
    + i.x + vec3(0.0, i1.x, 1.0));
  vec3 m = max(0.5 - vec3(dot(x0, x0), dot(x12.xy, x12.xy),
      dot(x12.zw, x12.zw)), 0.0);
  m = m * m;
  m = m * m;
  vec3 x = 2.0 * fract(p * C.www) - 1.0;
  vec3 h = abs(x) - 0.5;
  vec3 ox = floor(x + 0.5);
  vec3 a0 = x - ox;
  m *= 1.79284291400159 - 0.85373472095314 * (a0 * a0 + h * h);
  vec3 g;
  g.x = a0.x * x0.x + h.x * x0.y;
  g.yz = a0.yz * x12.xz + h.yz * x12.yw;
  return 130.0 * dot(m, g);
}
`,ve=`#version 300 es
precision mediump float;

uniform sampler2D u_image;
uniform float u_imageAspectRatio;

uniform vec2 u_resolution;
uniform float u_time;

uniform vec4 u_colorBack;
uniform vec4 u_colorTint;

uniform float u_softness;
uniform float u_repetition;
uniform float u_shiftRed;
uniform float u_shiftBlue;
uniform float u_distortion;
uniform float u_contour;
uniform float u_angle;

uniform float u_shape;
uniform bool u_isImage;

in vec2 v_objectUV;
in vec2 v_responsiveUV;
in vec2 v_responsiveBoxGivenSize;
in vec2 v_imageUV;

out vec4 fragColor;

${fe}
${pe}
${ge}

float getColorChanges(float c1, float c2, float stripe_p, vec3 w, float blur, float bump, float tint) {

  float ch = mix(c2, c1, smoothstep(.0, 2. * blur, stripe_p));

  float border = w[0];
  ch = mix(ch, c2, smoothstep(border, border + 2. * blur, stripe_p));

  if (u_isImage == true) {
    bump = smoothstep(.2, .8, bump);
  }
  border = w[0] + .4 * (1. - bump) * w[1];
  ch = mix(ch, c1, smoothstep(border, border + 2. * blur, stripe_p));

  border = w[0] + .5 * (1. - bump) * w[1];
  ch = mix(ch, c2, smoothstep(border, border + 2. * blur, stripe_p));

  border = w[0] + w[1];
  ch = mix(ch, c1, smoothstep(border, border + 2. * blur, stripe_p));

  float gradient_t = (stripe_p - w[0] - w[1]) / w[2];
  float gradient = mix(c1, c2, smoothstep(0., 1., gradient_t));
  ch = mix(ch, gradient, smoothstep(border, border + .5 * blur, stripe_p));

  // Tint color is applied with color burn blending
  ch = mix(ch, 1. - min(1., (1. - ch) / max(tint, 0.0001)), u_colorTint.a);
  return ch;
}

float getImgFrame(vec2 uv, float th) {
  float frame = 1.;
  frame *= smoothstep(0., th, uv.y);
  frame *= 1.0 - smoothstep(1. - th, 1., uv.y);
  frame *= smoothstep(0., th, uv.x);
  frame *= 1.0 - smoothstep(1. - th, 1., uv.x);
  return frame;
}

float blurEdge3x3(sampler2D tex, vec2 uv, vec2 dudx, vec2 dudy, float radius, float centerSample) {
  vec2 texel = 1.0 / vec2(textureSize(tex, 0));
  vec2 r = radius * texel;

  float w1 = 1.0, w2 = 2.0, w4 = 4.0;
  float norm = 16.0;
  float sum = w4 * centerSample;

  sum += w2 * textureGrad(tex, uv + vec2(0.0, -r.y), dudx, dudy).r;
  sum += w2 * textureGrad(tex, uv + vec2(0.0, r.y), dudx, dudy).r;
  sum += w2 * textureGrad(tex, uv + vec2(-r.x, 0.0), dudx, dudy).r;
  sum += w2 * textureGrad(tex, uv + vec2(r.x, 0.0), dudx, dudy).r;

  sum += w1 * textureGrad(tex, uv + vec2(-r.x, -r.y), dudx, dudy).r;
  sum += w1 * textureGrad(tex, uv + vec2(r.x, -r.y), dudx, dudy).r;
  sum += w1 * textureGrad(tex, uv + vec2(-r.x, r.y), dudx, dudy).r;
  sum += w1 * textureGrad(tex, uv + vec2(r.x, r.y), dudx, dudy).r;

  return sum / norm;
}

float lst(float edge0, float edge1, float x) {
  return clamp((x - edge0) / (edge1 - edge0), 0.0, 1.0);
}

void main() {

  const float firstFrameOffset = 2.8;
  float t = .3 * (u_time + firstFrameOffset);

  vec2 uv = v_imageUV;
  vec2 dudx = dFdx(v_imageUV);
  vec2 dudy = dFdy(v_imageUV);
  vec4 img = textureGrad(u_image, uv, dudx, dudy);

  if (u_isImage == false) {
    uv = v_objectUV + .5;
    uv.y = 1. - uv.y;
  }

  float cycleWidth = u_repetition;
  float edge = 0.;
  float contOffset = 1.;

  vec2 rotatedUV = uv - vec2(.5);
  float angle = (-u_angle + 70.) * PI / 180.;
  float cosA = cos(angle);
  float sinA = sin(angle);
  rotatedUV = vec2(
  rotatedUV.x * cosA - rotatedUV.y * sinA,
  rotatedUV.x * sinA + rotatedUV.y * cosA
  ) + vec2(.5);

  if (u_isImage == true) {
    float edgeRaw = img.r;
    edge = blurEdge3x3(u_image, uv, dudx, dudy, 6., edgeRaw);
    edge = pow(edge, 1.6);
    edge *= mix(0.0, 1.0, smoothstep(0.0, 0.4, u_contour));
  } else {
    if (u_shape < 1.) {
      // full-fill on canvas
      vec2 borderUV = v_responsiveUV + .5;
      float ratio = v_responsiveBoxGivenSize.x / v_responsiveBoxGivenSize.y;
      vec2 mask = min(borderUV, 1. - borderUV);
      vec2 pixel_thickness = min(250. / v_responsiveBoxGivenSize, vec2(.5));
      float maskX = smoothstep(0.0, pixel_thickness.x, mask.x);
      float maskY = smoothstep(0.0, pixel_thickness.y, mask.y);
      maskX = pow(maskX, .25);
      maskY = pow(maskY, .25);
      edge = clamp(1. - maskX * maskY, 0., 1.);

      uv = v_responsiveUV;
      if (ratio > 1.) {
        uv.y /= ratio;
      } else {
        uv.x *= ratio;
      }
      uv += .5;
      uv.y = 1. - uv.y;

      cycleWidth *= 2.;
      contOffset = 1.5;

    } else if (u_shape < 2.) {
      // circle
      vec2 shapeUV = uv - .5;
      shapeUV *= .67;
      edge = pow(clamp(3. * length(shapeUV), 0., 1.), 18.);
    } else if (u_shape < 3.) {
      // daisy
      vec2 shapeUV = uv - .5;
      shapeUV *= 1.68;

      float r = length(shapeUV) * 2.;
      float a = atan(shapeUV.y, shapeUV.x) + .2;
      r *= (1. + .05 * sin(3. * a + 2. * t));
      float f = abs(cos(a * 3.));
      edge = smoothstep(f, f + .7, r);
      edge *= edge;

      uv *= .8;
      cycleWidth *= 1.6;

    } else if (u_shape < 4.) {
      // diamond
      vec2 shapeUV = uv - .5;
      shapeUV = rotate(shapeUV, .25 * PI);
      shapeUV *= 1.42;
      shapeUV += .5;
      vec2 mask = min(shapeUV, 1. - shapeUV);
      vec2 pixel_thickness = vec2(.15);
      float maskX = smoothstep(0.0, pixel_thickness.x, mask.x);
      float maskY = smoothstep(0.0, pixel_thickness.y, mask.y);
      maskX = pow(maskX, .25);
      maskY = pow(maskY, .25);
      edge = clamp(1. - maskX * maskY, 0., 1.);
    } else if (u_shape < 5.) {
      // metaballs
      vec2 shapeUV = uv - .5;
      shapeUV *= 1.3;
      edge = 0.;
      for (int i = 0; i < 5; i++) {
        float fi = float(i);
        float speed = 1.5 + 2./3. * sin(fi * 12.345);
        float angle = -fi * 1.5;
        vec2 dir1 = vec2(cos(angle), sin(angle));
        vec2 dir2 = vec2(cos(angle + 1.57), sin(angle + 1.));
        vec2 traj = .4 * (dir1 * sin(t * speed + fi * 1.23) + dir2 * cos(t * (speed * 0.7) + fi * 2.17));
        float d = length(shapeUV + traj);
        edge += pow(1.0 - clamp(d, 0.0, 1.0), 4.0);
      }
      edge = 1. - smoothstep(.65, .9, edge);
      edge = pow(edge, 4.);
    }

    edge = mix(smoothstep(.9 - 2. * fwidth(edge), .9, edge), edge, smoothstep(0.0, 0.4, u_contour));

  }

  float opacity = 0.;
  if (u_isImage == true) {
    opacity = img.g;
    float frame = getImgFrame(v_imageUV, 0.);
    opacity *= frame;
  } else {
    opacity = 1. - smoothstep(.9 - 2. * fwidth(edge), .9, edge);
    if (u_shape < 2.) {
      edge = 1.2 * edge;
    } else if (u_shape < 5.) {
      edge = 1.8 * pow(edge, 1.5);
    }
  }

  float diagBLtoTR = rotatedUV.x - rotatedUV.y;
  float diagTLtoBR = rotatedUV.x + rotatedUV.y;

  vec3 color = vec3(0.);
  vec3 color1 = vec3(.98, 0.98, 1.);
  vec3 color2 = vec3(.1, .1, .1 + .1 * smoothstep(.7, 1.3, diagTLtoBR));

  vec2 grad_uv = uv - .5;

  float dist = length(grad_uv + vec2(0., .2 * diagBLtoTR));
  grad_uv = rotate(grad_uv, (.25 - .2 * diagBLtoTR) * PI);
  float direction = grad_uv.x;

  float bump = pow(1.8 * dist, 1.2);
  bump = 1. - bump;
  bump *= pow(uv.y, .3);


  float thin_strip_1_ratio = .12 / cycleWidth * (1. - .4 * bump);
  float thin_strip_2_ratio = .07 / cycleWidth * (1. + .4 * bump);
  float wide_strip_ratio = (1. - thin_strip_1_ratio - thin_strip_2_ratio);

  float thin_strip_1_width = cycleWidth * thin_strip_1_ratio;
  float thin_strip_2_width = cycleWidth * thin_strip_2_ratio;

  float noise = snoise(uv - t);

  edge += (1. - edge) * u_distortion * noise;

  direction += diagBLtoTR;
  float contour = 0.;
  direction -= 2. * noise * diagBLtoTR * (smoothstep(0., 1., edge) * (1.0 - smoothstep(0., 1., edge)));
  direction *= mix(1., 1. - edge, smoothstep(.5, 1., u_contour));
  direction -= 1.7 * edge * smoothstep(.5, 1., u_contour);
  direction += .2 * pow(u_contour, 4.) * (1.0 - smoothstep(0., 1., edge));

  bump *= clamp(pow(uv.y, .1), .3, 1.);
  direction *= (.1 + (1.1 - edge) * bump);

  direction *= (.4 + .6 * (1.0 - smoothstep(.5, 1., edge)));
  direction += .18 * (smoothstep(.1, .2, uv.y) * (1.0 - smoothstep(.2, .4, uv.y)));
  direction += .03 * (smoothstep(.1, .2, 1. - uv.y) * (1.0 - smoothstep(.2, .4, 1. - uv.y)));

  direction *= (.5 + .5 * pow(uv.y, 2.));
  direction *= cycleWidth;
  direction -= t;


  float colorDispersion = (1. - bump);
  colorDispersion = clamp(colorDispersion, 0., 1.);
  float dispersionRed = colorDispersion;
  dispersionRed += .03 * bump * noise;
  dispersionRed += 5. * (smoothstep(-.1, .2, uv.y) * (1.0 - smoothstep(.1, .5, uv.y))) * (smoothstep(.4, .6, bump) * (1.0 - smoothstep(.4, 1., bump)));
  dispersionRed -= diagBLtoTR;

  float dispersionBlue = colorDispersion;
  dispersionBlue *= 1.3;
  dispersionBlue += (smoothstep(0., .4, uv.y) * (1.0 - smoothstep(.1, .8, uv.y))) * (smoothstep(.4, .6, bump) * (1.0 - smoothstep(.4, .8, bump)));
  dispersionBlue -= .2 * edge;

  dispersionRed *= (u_shiftRed / 20.);
  dispersionBlue *= (u_shiftBlue / 20.);

  float blur = 0.;
  float rExtraBlur = 0.;
  float gExtraBlur = 0.;
  if (u_isImage == true) {
    float softness = 0.05 * u_softness;
    blur = softness + .5 * smoothstep(1., 10., u_repetition) * smoothstep(.0, 1., edge);
    float smallCanvasT = 1.0 - smoothstep(100., 500., min(u_resolution.x, u_resolution.y));
    blur += smallCanvasT * smoothstep(.0, 1., edge);
    rExtraBlur = softness * (0.05 + .1 * (u_shiftRed / 20.) * bump);
    gExtraBlur = softness * 0.05 / max(0.001, abs(1. - diagBLtoTR));
  } else {
    blur = u_softness / 15. + .3 * contour;
  }

  vec3 w = vec3(thin_strip_1_width, thin_strip_2_width, wide_strip_ratio);
  w[1] -= .02 * smoothstep(.0, 1., edge + bump);
  float stripe_r = fract(direction + dispersionRed);
  float r = getColorChanges(color1.r, color2.r, stripe_r, w, blur + fwidth(stripe_r) + rExtraBlur, bump, u_colorTint.r);
  float stripe_g = fract(direction);
  float g = getColorChanges(color1.g, color2.g, stripe_g, w, blur + fwidth(stripe_g) + gExtraBlur, bump, u_colorTint.g);
  float stripe_b = fract(direction - dispersionBlue);
  float b = getColorChanges(color1.b, color2.b, stripe_b, w, blur + fwidth(stripe_b), bump, u_colorTint.b);

  color = vec3(r, g, b);
  color *= opacity;

  vec3 bgColor = u_colorBack.rgb * u_colorBack.a;
  color = color + bgColor * (1. - opacity);
  opacity = opacity + u_colorBack.a * (1. - opacity);

  ${me}

  fragColor = vec4(color, opacity);
}
`,ee={workingSize:512,iterations:40};function $(t){const e=document.createElement("canvas"),o=e.getContext("2d"),i=typeof t=="string"&&t.startsWith("blob:");return new Promise((n,s)=>{if(!t||!o){s(new Error("Invalid file or canvas context"));return}const a=i&&fetch(t).then(f=>f.headers.get("Content-Type")),r=new Image;r.crossOrigin="anonymous",performance.now(),r.onload=async()=>{let f;const c=await a;c?f=c==="image/svg+xml":typeof t=="string"?f=t.endsWith(".svg")||t.startsWith("data:image/svg+xml"):f=t.type==="image/svg+xml";let d=r.width||r.naturalWidth,h=r.height||r.naturalHeight;if(f){const b=d/h;d>h?(d=4096,h=4096/b):(h=4096,d=4096*b),r.width=d,r.height=h}const R=Math.min(d,h),v=ee.workingSize/R,u=Math.round(d*v),p=Math.round(h*v);e.width=d,e.height=h;const _=document.createElement("canvas");_.width=u,_.height=p;const B=_.getContext("2d");B.drawImage(r,0,0,u,p),performance.now();const x=B.getImageData(0,0,u,p).data,m=new Uint8Array(u*p),I=new Uint8Array(u*p);for(let g=0,b=0;g<x.length;g+=4,b++){const S=x[g+3]===0?0:1;m[b]=S}const M=[],L=[];for(let g=0;g<p;g++)for(let b=0;b<u;b++){const y=g*u+b;if(!m[y])continue;let S=!1;b===0||b===u-1||g===0||g===p-1?S=!0:S=!m[y-1]||!m[y+1]||!m[y-u]||!m[y+u]||!m[y-u-1]||!m[y-u+1]||!m[y+u-1]||!m[y+u+1],S?(I[y]=1,M.push(y)):L.push(y)}const T=xe(m,I,new Uint32Array(L),new Uint32Array(M),u,p);performance.now();const F=_e(T,m,I,u,p);let C=0,D;for(let g=0;g<L.length;g++){const b=L[g];F[b]>C&&(C=F[b])}const O=document.createElement("canvas");O.width=u,O.height=p;const P=O.getContext("2d"),E=P.createImageData(u,p);for(let g=0;g<p;g++)for(let b=0;b<u;b++){const y=g*u+b,S=y*4;if(!m[y])E.data[S]=255,E.data[S+1]=255,E.data[S+2]=255,E.data[S+3]=0;else{const H=255*(1-F[y]/C);E.data[S]=H,E.data[S+1]=H,E.data[S+2]=H,E.data[S+3]=255}}P.putImageData(E,0,0),o.imageSmoothingEnabled=!0,o.imageSmoothingQuality="high",o.drawImage(O,0,0,u,p,0,0,d,h);const A=o.getImageData(0,0,d,h),k=document.createElement("canvas");k.width=d,k.height=h;const j=k.getContext("2d");j.drawImage(r,0,0,d,h);const ie=j.getImageData(0,0,d,h);for(let g=0;g<A.data.length;g+=4){const b=ie.data[g+3],y=A.data[g+3];b===0?(A.data[g]=255,A.data[g+1]=0):(A.data[g]=y===0?0:A.data[g],A.data[g+1]=b),A.data[g+2]=255,A.data[g+3]=255}o.putImageData(A,0,0),D=A,e.toBlob(g=>{if(!g){s(new Error("Failed to create PNG blob"));return}n({imageData:D,pngBlob:g})},"image/png")},r.onerror=()=>s(new Error("Failed to load image")),r.src=typeof t=="string"?t:URL.createObjectURL(t)})}function xe(t,e,o,i,n,s){const a=o.length,r=new Int32Array(a*4);for(let f=0;f<a;f++){const c=o[f],d=c%n,h=Math.floor(c/n);r[f*4+0]=d<n-1&&t[c+1]?c+1:-1,r[f*4+1]=d>0&&t[c-1]?c-1:-1,r[f*4+2]=h>0&&t[c-n]?c-n:-1,r[f*4+3]=h<s-1&&t[c+n]?c+n:-1}return{interiorPixels:o,boundaryPixels:i,pixelCount:a,neighborIndices:r}}function _e(t,e,o,i,n){const s=ee.iterations,a=.01,r=new Float32Array(i*n),{interiorPixels:f,neighborIndices:c,pixelCount:d}=t;performance.now();const h=1.9,R=[],V=[];for(let v=0;v<d;v++){const u=f[v],p=u%i,_=Math.floor(u/i);(p+_)%2===0?R.push(v):V.push(v)}for(let v=0;v<s;v++){for(const u of R){const p=f[u],_=c[u*4+0],B=c[u*4+1],z=c[u*4+2],x=c[u*4+3];let m=0;_>=0&&(m+=r[_]),B>=0&&(m+=r[B]),z>=0&&(m+=r[z]),x>=0&&(m+=r[x]);const I=(a+m)/4;r[p]=h*I+(1-h)*r[p]}for(const u of V){const p=f[u],_=c[u*4+0],B=c[u*4+1],z=c[u*4+2],x=c[u*4+3];let m=0;_>=0&&(m+=r[_]),B>=0&&(m+=r[B]),z>=0&&(m+=r[z]),x>=0&&(m+=r[x]);const I=(a+m)/4;r[p]=h*I+(1-h)*r[p]}}return r}const be={none:0,circle:1,daisy:2,diamond:3,metaballs:4};function q(t){if(Array.isArray(t))return t.length===4?t:t.length===3?[...t,1]:X;if(typeof t!="string")return X;let e,o,i,n=1;if(t.startsWith("#"))[e,o,i,n]=we(t);else if(t.startsWith("rgb"))[e,o,i,n]=ye(t);else if(t.startsWith("hsl"))[e,o,i,n]=Ue(Re(t));else return console.error("Unsupported color format",t),X;return[W(e,0,1),W(o,0,1),W(i,0,1),W(n,0,1)]}function we(t){t=t.replace(/^#/,""),t.length===3&&(t=t.split("").map(s=>s+s).join("")),t.length===6&&(t=t+"ff");const e=parseInt(t.slice(0,2),16)/255,o=parseInt(t.slice(2,4),16)/255,i=parseInt(t.slice(4,6),16)/255,n=parseInt(t.slice(6,8),16)/255;return[e,o,i,n]}function ye(t){const e=t.match(/^rgba?\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([0-9.]+))?\s*\)$/i);return e?[parseInt(e[1]??"0")/255,parseInt(e[2]??"0")/255,parseInt(e[3]??"0")/255,e[4]===void 0?1:parseFloat(e[4])]:[0,0,0,1]}function Re(t){const e=t.match(/^hsla?\s*\(\s*(\d+)\s*,\s*(\d+)%\s*,\s*(\d+)%\s*(?:,\s*([0-9.]+))?\s*\)$/i);return e?[parseInt(e[1]??"0"),parseInt(e[2]??"0"),parseInt(e[3]??"0"),e[4]===void 0?1:parseFloat(e[4])]:[0,0,0,1]}function Ue(t){const[e,o,i,n]=t,s=e/360,a=o/100,r=i/100;let f,c,d;if(o===0)f=c=d=r;else{const h=(v,u,p)=>(p<0&&(p+=1),p>1&&(p-=1),p<.16666666666666666?v+(u-v)*6*p:p<.5?u:p<.6666666666666666?v+(u-v)*(.6666666666666666-p)*6:v),R=r<.5?r*(1+a):r+a-r*a,V=2*r-R;f=h(V,R,s+1/3),c=h(V,R,s),d=h(V,R,s-1/3)}return[f,c,d,n]}const W=(t,e,o)=>Math.min(Math.max(t,e),o),X=[0,0,0,1],Se="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==";function Ae(t){const e=U.useRef(void 0),o=U.useCallback(i=>{const n=t.map(s=>{if(s!=null){if(typeof s=="function"){const a=s,r=a(i);return typeof r=="function"?r:()=>{a(null)}}return s.current=i,()=>{s.current=null}}});return()=>{n.forEach(s=>s==null?void 0:s())}},t);return U.useMemo(()=>t.every(i=>i==null)?null:i=>{e.current&&(e.current(),e.current=void 0),i!=null&&(e.current=o(i))},t)}function K(t){if(t.naturalWidth<1024&&t.naturalHeight<1024){if(t.naturalWidth<1||t.naturalHeight<1)return;const e=t.naturalWidth/t.naturalHeight;t.width=Math.round(e>1?1024*e:1024),t.height=Math.round(e>1?1024:1024/e)}}async function Q(t){const e={},o=[],i=s=>{try{return s.startsWith("/")||new URL(s),!0}catch{return!1}},n=s=>{try{return s.startsWith("/")?!1:new URL(s,window.location.origin).origin!==window.location.origin}catch{return!1}};return Object.entries(t).forEach(([s,a])=>{if(typeof a=="string"){const r=a||Se;if(!i(r)){console.warn(`Uniform "${s}" has invalid URL "${r}". Skipping image loading.`);return}const f=new Promise((c,d)=>{const h=new Image;n(r)&&(h.crossOrigin="anonymous"),h.onload=()=>{K(h),e[s]=h,c()},h.onerror=()=>{console.error(`Could not set uniforms. Failed to load image at ${r}`),d()},h.src=r});o.push(f)}else a instanceof HTMLImageElement&&K(a),e[s]=a}),await Promise.all(o),e}const te=U.forwardRef(function({fragmentShader:e,uniforms:o,webGlContextAttributes:i,speed:n=0,frame:s=0,width:a,height:r,minPixelRatio:f,maxPixelCount:c,mipmaps:d,style:h,...R},V){const[v,u]=U.useState(!1),p=U.useRef(null),_=U.useRef(null),B=U.useRef(i);U.useEffect(()=>((async()=>{const m=await Q(o);p.current&&!_.current&&(_.current=new ae(p.current,e,m,B.current,n,s,f,c,d),u(!0))})(),()=>{var m;(m=_.current)==null||m.dispose(),_.current=null}),[e]),U.useEffect(()=>{let x=!1;return(async()=>{var M;const I=await Q(o);x||(M=_.current)==null||M.setUniforms(I)})(),()=>{x=!0}},[o,v]),U.useEffect(()=>{var x;(x=_.current)==null||x.setSpeed(n)},[n,v]),U.useEffect(()=>{var x;(x=_.current)==null||x.setMaxPixelCount(c)},[c,v]),U.useEffect(()=>{var x;(x=_.current)==null||x.setMinPixelRatio(f)},[f,v]),U.useEffect(()=>{var x;(x=_.current)==null||x.setFrame(s)},[s,v]);const z=Ae([p,V]);return G.jsx("div",{ref:z,style:a!==void 0||r!==void 0?{width:typeof a=="string"&&isNaN(+a)===!1?+a:a,height:typeof r=="string"&&isNaN(+r)===!1?+r:r,...h}:h,...R})});te.displayName="ShaderMount";const Z="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==",Be=t=>typeof t=="object"&&typeof t.then=="function",J=[];function Ee(t,e){if(t===e)return!0;if(!t||!e)return!1;const o=t.length;if(e.length!==o)return!1;for(let i=0;i<o;i++)if(t[i]!==e[i])return!1;return!0}function Ve(t,e=null){e===null&&(e=[t]);for(const i of J)if(Ee(e,i.keys)){if(Object.prototype.hasOwnProperty.call(i,"error"))throw i.error;if(Object.prototype.hasOwnProperty.call(i,"response"))return i.response;throw i.promise}const o={keys:e,promise:(Be(t)?t:t(...e)).then(i=>{o.response=i}).catch(i=>o.error=i)};throw J.push(o),o.promise}const Ie=(t,e)=>Ve(t,e),w={params:{...he,scale:.6,speed:1,frame:0,colorBack:"#AAAAAC",colorTint:"#ffffff",distortion:.07,repetition:2,shiftRed:.3,shiftBlue:.3,contour:.4,softness:.1,angle:70,shape:"diamond"}},ze=U.memo(function({colorBack:e=w.params.colorBack,colorTint:o=w.params.colorTint,speed:i=w.params.speed,frame:n=w.params.frame,image:s="",contour:a=w.params.contour,distortion:r=w.params.distortion,softness:f=w.params.softness,repetition:c=w.params.repetition,shiftRed:d=w.params.shiftRed,shiftBlue:h=w.params.shiftBlue,angle:R=w.params.angle,shape:V=w.params.shape,suspendWhenProcessingImage:v=!1,fit:u=w.params.fit,scale:p=w.params.scale,rotation:_=w.params.rotation,originX:B=w.params.originX,originY:z=w.params.originY,offsetX:x=w.params.offsetX,offsetY:m=w.params.offsetY,worldWidth:I=w.params.worldWidth,worldHeight:M=w.params.worldHeight,...L}){const T=typeof s=="string"?s:s.src,[F,C]=U.useState(Z);let D;v&&typeof window<"u"&&T?D=Ie(()=>$(T).then(P=>URL.createObjectURL(P.pngBlob)),[T,"liquid-metal"]):D=F,U.useLayoutEffect(()=>{if(v)return;if(!T){C(Z);return}let P,E=!0;return $(T).then(A=>{E&&(P=URL.createObjectURL(A.pngBlob),C(P))}),()=>{E=!1}},[T,v]);const O={u_colorBack:q(e),u_colorTint:q(o),u_image:D,u_contour:a,u_distortion:r,u_softness:f,u_repetition:c,u_shiftRed:d,u_shiftBlue:h,u_angle:R,u_isImage:!!s,u_shape:be[V],u_fit:de[u],u_scale:p,u_rotation:_,u_offsetX:x,u_offsetY:m,u_originX:B,u_originY:z,u_worldWidth:I,u_worldHeight:M};return G.jsx(te,{...L,speed:i,frame:n,fragmentShader:ve,mipmaps:["u_image"],uniforms:O})});function Ce({theme:t="dark",className:e="",scale:o=.6}){const i=t==="light";return G.jsx("div",{className:`superopc-liquid-metal ${e}`.trim(),role:"img","aria-label":"一战晟铭",children:G.jsx(ze,{image:i?"/assets/brand-logo-black.png":"/assets/brand-logo-white.png",speed:.65,scale:o,rotation:0,offsetX:0,offsetY:0,fit:"contain",colorBack:"rgba(0, 0, 0, 0)",colorTint:i?"#101218":"#f5f7fb",contour:.4,distortion:.07,softness:.1,repetition:2,shiftRed:.3,shiftBlue:.3,angle:70,maxPixelCount:18e4,webGlContextAttributes:{alpha:!0,preserveDrawingBuffer:!1},style:{width:"100%",height:"100%"}})})}export{Ce as default};

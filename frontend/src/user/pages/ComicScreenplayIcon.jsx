import { createElement, useId } from 'react';
import { originalComicIcons, originalComicProviders, originalComicProviderIndex } from './comic-screenplay-icon-data.js';

const tags = new Set(['svg', 'path', 'rect', 'circle', 'ellipse', 'g', 'line', 'polyline', 'polygon', 'span', 'defs', 'linearGradient', 'stop']);
const attributes = new Set(['xmlns', 'xmlns:xlink', 'viewBox', 'd', 'fill', 'stroke', 'stroke-miterlimit', 'stroke-width', 'stroke-linecap', 'stroke-linejoin', 'x', 'y', 'x1', 'x2', 'y1', 'y2', 'width', 'height', 'rx', 'ry', 'cx', 'cy', 'r', 'transform', 'points', 'opacity', 'fill-rule', 'clip-rule', 'id', 'gradientUnits', 'offset', 'stop-color', 'stop-opacity', 'class']);
const reactAttributes = { 'xmlns:xlink': 'xmlnsXlink', 'stroke-miterlimit': 'strokeMiterlimit', 'stroke-width': 'strokeWidth', 'stroke-linecap': 'strokeLinecap', 'stroke-linejoin': 'strokeLinejoin', 'fill-rule': 'fillRule', 'clip-rule': 'clipRule', 'stop-color': 'stopColor', 'stop-opacity': 'stopOpacity', class: 'className' };
function safeStyle(value) {
  const allowed = { background: 'background', 'border-radius': 'borderRadius', padding: 'padding', color: 'color' };
  return Object.fromEntries(String(value || '').split(';').map(rule => rule.split(':')).filter(([key,value]) => allowed[key?.trim()] && value && !/url\s*\(|expression|javascript:/i.test(value)).map(([key,value])=>[allowed[key.trim()],value.trim()]));
}
function safeProps(attrs = {}, ids = {}) {
  const props=Object.fromEntries(Object.entries(attrs).filter(([key]) => attributes.has(key)).map(([key, value]) => [reactAttributes[key] || key, key === 'id' ? ids[value] || value : typeof value === 'string' ? value.replace(/url\(#([^)]+)\)/g,(_,id)=>`url(#${ids[id]||id})`) : value]));
  if(attrs.style)props.style=safeStyle(attrs.style);
  return props;
}
function node(tree, key, ids = {}) {
  if (!tree || !tags.has(tree.tag)) return null;
  return createElement(tree.tag, { ...safeProps(tree.attrs,ids), key }, (tree.children || []).map((child, index) => node(child, index,ids)));
}
function collectIds(tree,prefix,result={}) { if(tree.attrs?.id)result[tree.attrs.id]=`${prefix}-${tree.attrs.id}`;for(const child of tree.children||[])collectIds(child,prefix,result);return result; }

// Geometry and stroke attributes come from the approved static catalog.
// Labels and handlers remain on the existing enclosing control.
export default function ComicScreenplayIcon({ name, size = 16, className = '', style }) {
  const definition = originalComicIcons[name];
  if (!definition) return null;
  const tree = definition.svg;
  return createElement('svg', {
    ...safeProps(tree.attrs), width: size, height: size, fill: tree.attrs?.fill || 'currentColor',
    className: `comic-original-icon${name==='InlineAssetSpinner'?' comic-original-spinner':''}${className ? ` ${className}` : ''}`,
    style: { display: 'inline-block', verticalAlign: '-0.125em', flexShrink: 0, ...style },
    'aria-hidden': 'true', focusable: 'false', 'data-original-icon': name,
  }, (tree.children || []).map((child, index) => node(child, index)));
}

export function ComicScreenplayModelIcon({ modelId, size = 18 }) {
  const prefix=`comic-provider-${useId().replace(/[^A-Za-z0-9_-]/g,'')}`;
  const index=originalComicProviderIndex(modelId),provider=originalComicProviders[index];
  const ids={};for(const tree of provider.tree)collectIds(tree,prefix,ids);
  return createElement('span',{className:'comic-original-provider',style:{width:size,height:size,fontSize:size,'--comic-provider-size':`${size}px`},'data-size':size,'aria-hidden':'true','data-original-provider':index},provider.tree.map((tree,index)=>node(tree,index,ids)));
}

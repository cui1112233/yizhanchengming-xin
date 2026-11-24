declare module '*pluginCatalog.js' {export const PLUGIN_CATALOG:any[];}
declare module '*localToolLogic.js' {export function formatText(input:string,options?:any):any;export function mergeDocuments(files:any[],options?:any):any;export function mergeSubtitles(files:any[],options?:any):any;}
declare module '*storyboardOrganizer.js' {export function parseStoryboardFiles(files:any[],options?:any):any[];export function buildStoryboardJson(shots:any[],options?:any):any;}

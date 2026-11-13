import test from 'node:test';
import assert from 'node:assert/strict';
import { installNotificationPreferencesSync } from './notificationPreferencesSync.js';

test('saved sound settings reach another tab, remain owner scoped and tolerate bad storage',()=>{
 const a=new EventTarget(),b=new EventTarget(),received=[];
 const event=(type,props)=>Object.assign(new Event(type),props);
 const stopA=installNotificationPreferencesSync({target:a,getOwner:()=> 'alice',onUpdate:()=>{},storage:{setItem:(key,newValue)=>b.dispatchEvent(event('storage',{key,newValue}))}});
 const stopB=installNotificationPreferencesSync({target:b,getOwner:()=> 'alice',onUpdate:p=>received.push(p),storage:{setItem(){}}});
 a.dispatchEvent(event('qiantie:notifications-updated',{detail:{soundEnabled:false,soundVolume:20,petVisible:true}}));
 assert.deepEqual(received,[{soundEnabled:false,soundVolume:20}]);
 b.dispatchEvent(event('storage',{key:'qiantie:notification-preferences:bob',newValue:'{}'}));
 b.dispatchEvent(event('storage',{key:'qiantie:notification-preferences:alice',newValue:'broken'}));
 assert.equal(received.length,1);stopA();stopB();
});

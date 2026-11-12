import test from 'node:test';
import assert from 'node:assert/strict';
import { libraryProjects, projectEditorUrl } from './projectLibrary.js';

test('作品库搜索和排序不修改原始项目列表', () => {
  const projects = [{ id: 'a', title: '春日 AI', updatedAt: '2026-11-10' }, { id: 'b', title: '夜景', updatedAt: '2026-11-12' }];
  assert.deepEqual(libraryProjects(projects).map(item => item.id), ['b', 'a']);
  assert.deepEqual(libraryProjects(projects, ' ai ').map(item => item.id), ['a']);
  assert.deepEqual(libraryProjects(projects, '不存在'), []);
  assert.equal(projects[0].id, 'a');
});

test('每个工程打开独立地址，刷新后仍定位原工程', () => {
  assert.equal(projectEditorUrl('project-123'), '/ai-cut?project=project-123');
  assert.equal(projectEditorUrl('a&b'), '/ai-cut?project=a%26b');
});
